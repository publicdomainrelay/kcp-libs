package dns

import (
	"context"
	"encoding/json"
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

	Paths PathResolver

	Domain string

	TokenTTL time.Duration

	ServiceAccountNamespace string
}

type DNS struct {
	opts Options
}

func New(opts Options) *DNS {
	if opts.Domain == "" {
		opts.Domain = kcp.DefaultServiceDomain
	}
	if opts.ServiceAccountNamespace == "" {
		opts.ServiceAccountNamespace = "default"
	}
	return &DNS{opts: opts}
}

func (d *DNS) Domain() string {
	return d.opts.Domain
}

func (d *DNS) Name(name, namespace, logicalCluster string) string {
	cluster := logicalCluster
	if d.opts.Paths != nil {
		if path := d.opts.Paths.Lookup(context.Background(), logicalCluster); path != "" {
			cluster = path
		}
	}
	return kcp.ServiceFQDN(name, namespace, cluster, d.opts.Domain)
}

func (d *DNS) Table() (map[string]string, []string) {
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
		table[d.Name(pod.GetName(), pod.GetNamespace(), cluster)] = address
	}
	return table, workspaces
}

func (d *DNS) Tokens(workspaces []string, account *denospec.ServiceAccountRef) string {
	out := map[string]string{}
	if account != nil && d.opts.Minter != nil {
		namespace := account.Namespace
		if namespace == "" {
			namespace = d.opts.ServiceAccountNamespace
		}
		for _, cluster := range workspaces {
			token, err := d.opts.Minter.MintServiceAccountToken(context.Background(), cluster, namespace, account.Name, d.opts.TokenTTL)
			if err != nil {
				continue
			}
			out[cluster] = token
		}
	}
	body, err := json.Marshal(out)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func (d *DNS) Env(target ref.Ref, selfArgs, selfEnv string, account *denospec.ServiceAccountRef) map[string]string {
	env := map[string]string{
		DomainKey:    d.opts.Domain,
		NamespaceKey: target.Namespace,
	}
	table, workspaces := d.Table()
	if self := AdvertisedAddress(selfArgs, selfEnv); self != "" {
		if !contains(workspaces, target.LogicalCluster) {
			workspaces = append(workspaces, target.LogicalCluster)
		}
		table[d.Name(target.Name, target.Namespace, target.LogicalCluster)] = self
	}
	if len(table) > 0 {
		if body, err := json.Marshal(table); err == nil {
			env[TableKey] = string(body)
		}
	}
	if account != nil {
		env[TokensKey] = d.Tokens(workspaces, account)
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

func contains(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}
