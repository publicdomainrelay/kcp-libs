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
	"k8s.io/apimachinery/pkg/runtime/schema"

	abcstore "github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/denospec"
	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/factory/dns"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

const (
	namespace = "default"

	domain = "kcp.local"

	serviceAccount = "default"

	firstPod = "pds"

	secondPod = "api"

	thirdPod = "pds"
)

var widgets = schema.GroupVersionResource{Group: livekcp.Group, Version: livekcp.Version, Resource: livekcp.Resource}

type metadata struct {
	Name string `json:"name"`

	Namespace string `json:"namespace"`

	Labels map[string]string `json:"labels,omitempty"`
}

type widget struct {
	APIVersion string `json:"apiVersion,omitempty"`

	Kind string `json:"kind,omitempty"`

	Metadata metadata `json:"metadata"`

	Spec struct {
		Env map[string]string `json:"env,omitempty"`
	} `json:"spec"`
}

func Run(ctx context.Context, out io.Writer) error {
	cluster, err := livekcp.Start(ctx)
	if err != nil {
		return err
	}
	defer cluster.Stop()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.Server, RestConfig: cluster.Config})
	if err != nil {
		return err
	}
	resource := kcpstore.Of[widget](store, widgets)

	if err := seed(ctx, resource, cluster.Consumer, firstPod, map[string]string{dns.ArgsKey: `["--port","8080"]`}); err != nil {
		return err
	}
	if err := seed(ctx, resource, cluster.Consumer, secondPod, map[string]string{dns.EnvKey: `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`}); err != nil {
		return err
	}
	if err := seed(ctx, resource, cluster.ConsumerTwo, thirdPod, map[string]string{dns.EnvKey: `{"PORT":"9000"}`}); err != nil {
		return err
	}

	workspaces := []string{cluster.Consumer, cluster.ConsumerTwo}
	source := dns.SourceFunc(func() []*unstructured.Unstructured {
		var out []*unstructured.Unstructured
		for _, logicalCluster := range workspaces {
			listed, err := resource.List(context.Background(), logicalCluster)
			if err != nil {
				continue
			}
			for i := range listed {
				out = append(out, unstructuredOf(logicalCluster, &listed[i]))
			}
		}
		return out
	})

	var _ abcstore.TokenMinter = store
	names := dns.New(dns.Options{
		Source:   source,
		Minter:   store,
		Paths:    kcpstore.NewPathCache(store),
		Domain:   domain,
		TokenTTL: time.Hour,
	})

	table, seen := names.Table()
	keys := make([]string, 0, len(table))
	for name := range table {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(out, "%s resolves to %s\n", name, table[name])
	}
	fmt.Fprintf(out, "the table covers %d workspaces\n", len(seen))

	self := ref.New(cluster.Consumer, namespace, firstPod)
	env := names.Env(self, `["--hostname","0.0.0.0","--port","8080"]`, "",
		&denospec.ServiceAccountRef{Name: serviceAccount})
	injected := map[string]string{}
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
	fmt.Fprintf(out, "the pod carries %d names and a token for %d workspaces\n", len(injected), len(tokenKeys))
	fmt.Fprintf(out, "its own name is in the table from its first moment: %s\n",
		injected[names.Name(firstPod, namespace, cluster.Consumer)])
	fmt.Fprintf(out, "kcp minted a token for %s: %v\n", cluster.Consumer, tokens[cluster.Consumer] != "")
	fmt.Fprintf(out, "a service that binds 0.0.0.0 advertises %s\n",
		dns.AdvertisedAddress("", `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`))
	fmt.Fprintf(out, "the workspace path reads as labels: %s becomes %s\n",
		cluster.Consumer, kcp.ServiceLabels(cluster.Consumer))
	return nil
}

func seed(ctx context.Context, resource *kcpstore.Resource[widget], cluster, name string, env map[string]string) error {
	_ = resource.Delete(ctx, ref.New(cluster, namespace, name))
	obj := &widget{APIVersion: livekcp.APIVersion, Kind: livekcp.Kind}
	obj.Metadata.Name = name
	obj.Metadata.Namespace = namespace
	obj.Spec.Env = env
	return resource.Create(ctx, cluster, obj)
}

func unstructuredOf(logicalCluster string, obj *widget) *unstructured.Unstructured {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	fields := map[string]any{}
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil
	}
	out := &unstructured.Unstructured{Object: fields}
	out.SetAnnotations(map[string]string{kcp.ClusterAnnotation: logicalCluster})
	return out
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-dns:", err)
		os.Exit(1)
	}
}
