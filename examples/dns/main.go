package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	abcstore "github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/denospec"
	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/factory/servicenames"
	"github.com/publicdomainrelay/kcp-libs/impl/assets"
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

type podSpec struct {
	Env map[string]string `json:"env,omitempty"`
}

type podStatus struct{}

type pod = livekcp.Object[podSpec, podStatus]

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
	resource := kcpstore.Of[pod](store, livekcp.WidgetGVR)

	if err := seed(ctx, cluster.ConsumerCluster, resource, firstPod, map[string]string{servicenames.ArgsKey: `["--port","8080"]`}); err != nil {
		return err
	}
	if err := seed(ctx, cluster.ConsumerCluster, resource, secondPod, map[string]string{servicenames.EnvKey: `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`}); err != nil {
		return err
	}
	if err := seed(ctx, cluster.SecondConsumerCluster, resource, thirdPod, map[string]string{servicenames.EnvKey: `{"PORT":"9000"}`}); err != nil {
		return err
	}

	workspaces := []string{cluster.ConsumerCluster, cluster.SecondConsumerCluster}
	source := servicenames.SourceFunc(func() []*unstructured.Unstructured {
		var out []*unstructured.Unstructured
		for _, logicalCluster := range workspaces {
			listed, err := resource.List(ctx, logicalCluster)
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
	resolver := servicenames.New(servicenames.Options{
		Source:   source,
		Minter:   store,
		Paths:    kcpstore.NewPathCache(store),
		Domain:   domain,
		TokenTTL: time.Hour,
	})

	table, seen := resolver.Table(ctx)
	keys := make([]string, 0, len(table))
	for name := range table {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(out, "%s resolves to %s\n", name, table[name])
	}
	fmt.Fprintf(out, "the table covers %d workspaces\n", len(seen))

	self := ref.New(cluster.ConsumerCluster, namespace, firstPod)
	env := resolver.Env(ctx, self, `["--hostname","0.0.0.0","--port","8080"]`, "",
		&denospec.ServiceAccountRef{Name: serviceAccount})
	injected := map[string]string{}
	if err := json.Unmarshal([]byte(env[servicenames.TableKey]), &injected); err != nil {
		return err
	}
	tokens := map[string]string{}
	if err := json.Unmarshal([]byte(env[servicenames.TokensKey]), &tokens); err != nil {
		return err
	}
	tokenKeys := make([]string, 0, len(tokens))
	for name := range tokens {
		tokenKeys = append(tokenKeys, name)
	}
	sort.Strings(tokenKeys)
	fmt.Fprintf(out, "the pod carries %d names and a token for %d workspaces\n", len(injected), len(tokenKeys))
	fmt.Fprintf(out, "its own name is in the table from its first moment: %s\n",
		injected[resolver.Name(ctx, firstPod, namespace, cluster.ConsumerCluster)])
	fmt.Fprintf(out, "kcp minted a token for %s: %v\n", cluster.ConsumerCluster, tokens[cluster.ConsumerCluster] != "")

	dir, err := os.MkdirTemp("", "kcpdns")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	paths, err := assets.DNSSet(dir).Materialise()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "the shim and probe are written under %s: %s and %s\n",
		assets.DNSDirName, filepath.Base(paths[assets.ShimName]), filepath.Base(paths[assets.ProbeName]))
	fmt.Fprintf(out, "a readiness probe preloads %s and is granted %s\n",
		assets.ShimName, strings.Join(assets.DNSProbeEnv, ","))
	fmt.Fprintf(out, "a service that binds 0.0.0.0 advertises %s\n",
		servicenames.AdvertisedAddress("", `{"PORT":"3000","HOSTNAME":"0.0.0.0"}`))
	fmt.Fprintf(out, "the workspace path reads as labels: %s becomes %s\n",
		cluster.ConsumerCluster, kcp.ServiceLabels(cluster.ConsumerCluster))
	return nil
}

func seed(ctx context.Context, cluster string, resource *kcpstore.Resource[pod], name string, env map[string]string) error {
	obj := livekcp.NewObject[podSpec, podStatus](namespace, name)
	obj.Spec.Env = env
	return livekcp.Seed(ctx, cluster, resource, namespace, name, obj)
}

func unstructuredOf(logicalCluster string, obj *pod) *unstructured.Unstructured {
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
