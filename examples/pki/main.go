package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/publicdomainrelay/kcp-libs/abc/pki"
	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/impl/openbaoclient"
	"github.com/publicdomainrelay/kcp-libs/impl/pkiprovisioner"
)

const (
	domain = "kcp.local"

	rootNamespace = "root"

	namespace = "alice.default"

	workspace = "root:alice"
)

func Run(ctx context.Context, out io.Writer) error {
	vault, err := newVault()
	if err != nil {
		return err
	}
	defer vault.Close()

	client, err := openbaoclient.New(openbaoclient.Options{Address: vault.URL(), Token: "root-token"})
	if err != nil {
		return err
	}
	var _ pki.Client = client
	health, err := client.Health(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "vault reachable, sealed %v\n", health.Sealed)

	provisioner, err := pkiprovisioner.New(pkiprovisioner.Options{
		Client:          client,
		RootNamespace:   rootNamespace,
		Domain:          domain,
		LeafTTL:         "720h",
		IntermediateTTL: "43800h",
	})
	if err != nil {
		return err
	}
	var _ pki.Provisioner = provisioner
	authority, err := provisioner.EnsureAuthority(ctx, namespace)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "intermediate %s serial %s\n", authority.CommonName, authority.Serial)

	fqdn := kcp.ServiceFQDN("pds", "default", workspace, domain)
	cert, err := provisioner.Issue(ctx, namespace, fqdn, []string{fqdn}, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "issued %s serial %s\n", fqdn, cert.Serial)
	fmt.Fprintf(out, "the workload serves %d certificates: its leaf, the namespace intermediate, and the root\n",
		strings.Count(pkiprovisioner.LeafChain(cert), "BEGIN CERTIFICATE")+1)

	fmt.Fprintf(out, "the root a peer verifies against is %d bytes and never read from the vault\n", len(provisioner.CachedRootPEM()))

	before := vault.CallCount()
	if _, err := provisioner.Issue(ctx, namespace, fqdn, []string{fqdn}, nil); err != nil {
		return err
	}
	fmt.Fprintf(out, "a second leaf cost %d calls, the authority was cached\n", vault.CallCount()-before)

	if err := provisioner.Delete(ctx, namespace); err != nil {
		return err
	}
	after := vault.CallCount()
	if _, err := provisioner.EnsureAuthority(ctx, namespace); err != nil {
		return err
	}
	fmt.Fprintf(out, "provisioning again after a delete cost %d calls, the namespace was gone\n", vault.CallCount()-after)
	return nil
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-pki:", err)
		os.Exit(1)
	}
}
