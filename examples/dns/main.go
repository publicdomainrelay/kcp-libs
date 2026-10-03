package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"

	abcstore "github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/denospec"
	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/factory/dns"
	"github.com/publicdomainrelay/kcp-libs/fakekcp"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
)

const (
	apiVersion = "example.computer/v1alpha1"

	podKind = "DenoPod"

	podResource = "denopods"

	domain = "kcp.local"

	aliceID = "2j35eh7jjhsc8ny9"

	alicePath = "root:alice"

	bobID = "8fj2hd8jjdhs09ab"

	bobPath = "root:bob"
)

func Run(ctx context.Context, out io.Writer) error {
	cluster, err := fakekcp.New()
	if err != nil {
		return err
	}
	defer cluster.Close()

	cluster.SetClusterPath(aliceID, alicePath)
	cluster.SetClusterPath(bobID, bobPath)
	cluster.Create(aliceID, "default", podResource, pod("pds", aliceID, "default", `["--port","8080"]`, ""))
	cluster.Create(aliceID, "default", podResource, pod("api", aliceID, "default", "", `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`))
	cluster.Create(bobID, "default", podResource, pod("pds", bobID, "default", "", `{"PORT":"9000"}`))

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.URL(), RestConfig: &rest.Config{Host: cluster.URL()}})
	if err != nil {
		return err
	}
	var _ abcstore.TokenMinter = store
	names := dns.New(dns.Options{
		Source:   source(cluster),
		Minter:   store,
		Paths:    kcpstore.NewPathCache(store),
		Domain:   domain,
		TokenTTL: time.Hour,
	})

	table, workspaces := names.Table()
	keys := make([]string, 0, len(table))
	for name := range table {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(out, "%s resolves to %s\n", name, table[name])
	}
	fmt.Fprintf(out, "the table covers %d workspaces\n", len(workspaces))

	self := ref.New(aliceID, "default", "pds")
	env := names.Env(self, `["--hostname","0.0.0.0","--port","8080"]`, "", &denospec.ServiceAccountRef{Name: "reader"})
	var injected map[string]string
	if err := json.Unmarshal([]byte(env[dns.TableKey]), &injected); err != nil {
		return err
	}
	tokens := map[string]string{}
	if err := json.Unmarshal([]byte(env[dns.TokensKey]), &tokens); err != nil {
		return err
	}
	tokenKeys := make([]string, 0, len(tokens))
	for name := range tokens {
		tokenKeys = append(tokenKeys, name)
	}
	sort.Strings(tokenKeys)
	fmt.Fprintf(out, "the pod carries %d names and a token for %v\n", len(injected), tokenKeys)
	fmt.Fprintf(out, "its own name is in the table from its first moment: %s\n", injected[names.Name("pds", "default", aliceID)])

	fmt.Fprintf(out, "a service that binds %s advertises %s\n", "0.0.0.0", dns.AdvertisedAddress("", `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`))
	fmt.Fprintf(out, "the workspace path reads as labels: %s becomes %s\n", alicePath, kcp.ServiceLabels(alicePath))
	return nil
}

func source(cluster *fakekcp.Cluster) dns.Source {
	return dns.SourceFunc(func() []*unstructured.Unstructured {
		var out []*unstructured.Unstructured
		for _, body := range cluster.List(fakekcp.Wildcard, podResource) {
			out = append(out, &unstructured.Unstructured{Object: body})
		}
		return out
	})
}

func pod(name, cluster, namespace, serviceArgs, serviceEnv string) map[string]any {
	body := fakekcp.Object(apiVersion, podKind, namespace, name)
	fakekcp.WithAnnotation(body, kcp.ClusterAnnotation, cluster)
	env := map[string]any{}
	if serviceArgs != "" {
		env[dns.ArgsKey] = serviceArgs
	}
	if serviceEnv != "" {
		env[dns.EnvKey] = serviceEnv
	}
	fakekcp.WithSpec(body, map[string]any{"env": env})
	return body
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-dns:", err)
		os.Exit(1)
	}
}
