package servicenames

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/denospec"
	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

const (
	ArgsKey = "SERVICE_ARGS"

	EnvKey = "SERVICE_ENV"

	TableKey = "KCP_DNS_TABLE"

	TokensKey = "KCP_TOKENS"

	DomainKey = "KCP_SERVICE_DOMAIN"

	NamespaceKey = "KCP_NAMESPACE"

	ShimKey = "KCP_SHIM"
)

type Source interface {
	Pods() []*unstructured.Unstructured
}

type SourceFunc func() []*unstructured.Unstructured

func (f SourceFunc) Pods() []*unstructured.Unstructured {
	return f()
}

type Workspaces interface {
	Clusters(ctx context.Context) []string
}

type WorkspacesFunc func(ctx context.Context) []string

func (f WorkspacesFunc) Clusters(ctx context.Context) []string {
	return f(ctx)
}

type PathResolver interface {
	Lookup(ctx context.Context, logicalCluster string) string
}

type PathResolverFunc func(ctx context.Context, logicalCluster string) string

func (f PathResolverFunc) Lookup(ctx context.Context, logicalCluster string) string {
	return f(ctx, logicalCluster)
}

type Options struct {
	Source Source

	Minter store.TokenMinter

	Workspaces Workspaces

	Paths PathResolver

	Domain string

	TokenTTL time.Duration

	ServiceAccountNamespace string

	Shim string
}

type Resolver struct {
	opts Options
}

func New(opts Options) *Resolver {
	if opts.Domain == "" {
		opts.Domain = kcp.DefaultServiceDomain
	}
	if opts.ServiceAccountNamespace == "" {
		opts.ServiceAccountNamespace = "default"
	}
	return &Resolver{opts: opts}
}

func (d *Resolver) clusterName(ctx context.Context, logicalCluster string) string {
	if d.opts.Paths != nil {
		if path := d.opts.Paths.Lookup(ctx, logicalCluster); path != "" {
			return path
		}
	}
	return logicalCluster
}

func (d *Resolver) Name(ctx context.Context, name, namespace, logicalCluster string) string {
	return kcp.ServiceFQDN(name, namespace, d.clusterName(ctx, logicalCluster), d.opts.Domain)
}

func (d *Resolver) Table(ctx context.Context) (map[string]string, []string) {
	table := map[string]string{}
	seen := map[string]bool{}
	var workspaces []string
	if d.opts.Source == nil {
		return table, workspaces
	}
	for _, pod := range d.opts.Source.Pods() {
		cluster := pod.GetAnnotations()[kcp.ClusterAnnotation]
		if cluster == "" {
			continue
		}
		if !seen[cluster] {
			seen[cluster] = true
			workspaces = append(workspaces, cluster)
		}
		address := AdvertisedAddress(envOf(pod, ArgsKey), envOf(pod, EnvKey))
		if address == "" {
			continue
		}
		table[d.Name(ctx, pod.GetName(), pod.GetNamespace(), cluster)] = address
	}
	return table, workspaces
}

func (d *Resolver) Tokens(ctx context.Context, workspaces []string, account *denospec.ServiceAccountRef) string {
	out := map[string]string{}
	if account != nil && d.opts.Minter != nil {
		namespace := account.Namespace
		if namespace == "" {
			namespace = d.opts.ServiceAccountNamespace
		}
		for _, workspace := range workspaces {
			key := d.clusterName(ctx, workspace)
			if _, done := out[key]; done {
				continue
			}
			token, err := d.opts.Minter.MintServiceAccountToken(ctx, workspace, namespace, account.Name, d.opts.TokenTTL)
			if err != nil {
				continue
			}
			out[key] = token
		}
	}
	body, err := json.Marshal(out)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func (d *Resolver) Env(ctx context.Context, target ref.Ref, selfArgs, selfEnv string, account *denospec.ServiceAccountRef) map[string]string {
	env := map[string]string{
		DomainKey:    d.opts.Domain,
		NamespaceKey: target.Namespace,
	}
	if d.opts.Shim != "" {
		env[ShimKey] = d.opts.Shim
	}
	table, workspaces := d.Table(ctx)
	if d.opts.Workspaces != nil {
		for _, cluster := range d.opts.Workspaces.Clusters(ctx) {
			if !slices.Contains(workspaces, cluster) {
				workspaces = append(workspaces, cluster)
			}
		}
	}
	if !slices.Contains(workspaces, target.LogicalCluster) {
		workspaces = append(workspaces, target.LogicalCluster)
	}
	if self := AdvertisedAddress(selfArgs, selfEnv); self != "" {
		table[d.Name(ctx, target.Name, target.Namespace, target.LogicalCluster)] = self
	}
	if len(table) > 0 {
		if body, err := json.Marshal(table); err == nil {
			env[TableKey] = string(body)
		}
	}
	if account != nil {
		env[TokensKey] = d.Tokens(ctx, workspaces, account)
	}
	return env
}

func AdvertisedAddress(argsJSON, envJSON string) string {
	args := []string{}
	if argsJSON != "" {
		_ = json.Unmarshal([]byte(argsJSON), &args)
	}
	serviceEnv := map[string]string{}
	if envJSON != "" {
		_ = json.Unmarshal([]byte(envJSON), &serviceEnv)
	}
	flag := func(name string) string {
		for i, arg := range args {
			if arg == name && i+1 < len(args) {
				return args[i+1]
			}
		}
		return ""
	}
	port := flag("--port")
	if port == "" {
		port = serviceEnv["PORT"]
	}
	if port == "" {
		return ""
	}
	host := flag("--hostname")
	if host == "" {
		host = serviceEnv["HOSTNAME"]
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return host + ":" + port
}

func envOf(pod *unstructured.Unstructured, key string) string {
	env, _, _ := unstructured.NestedMap(pod.Object, "spec", "env")
	value, _ := env[key].(string)
	return value
}
