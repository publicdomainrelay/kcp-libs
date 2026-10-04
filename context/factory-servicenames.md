# Context: factory-servicenames

Repository: `kcp-libs`

This context exists so that workloads running under kcp can discover each other by name without a hand-written DNS configuration. It converts the pods a cluster controller already observes into a service table (name to advertised host:port), the workspace list worth addressing, and the service account tokens needed to reach them, then hands that whole bundle to a workload as environment variables. It also defines the seams (Source, PathResolver, TokenMinter) that keep the package free of any particular client, so a caller supplies real store-backed pods and paths while tests supply plain functions.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: factory/servicenames/servicenames.go
  kind: function
  name: AdvertisedAddress
  signature: func AdvertisedAddress(argsJSON, envJSON string) string
- file: factory/servicenames/servicenames.go
  kind: function
  name: New
  signature: func New(opts Options) *Resolver
- file: factory/servicenames/servicenames.go
  kind: struct
  name: Options
  signature: type Options struct { Source Source; Minter store.TokenMinter; Paths
    PathResolver; Domain string; TokenTTL time.Duration; ServiceAccountNamespace string;
    Shim string }
- file: factory/servicenames/servicenames.go
  kind: interface
  name: PathResolver
  signature: type PathResolver interface { Lookup(ctx context.Context, logicalCluster
    string) string }
- file: factory/servicenames/servicenames.go
  kind: method
  name: PathResolver.Lookup
  signature: Lookup(ctx context.Context, logicalCluster string) string
- file: factory/servicenames/servicenames.go
  kind: type_alias
  name: PathResolverFunc
  signature: type PathResolverFunc func(ctx context.Context, logicalCluster string)
    string
- file: factory/servicenames/servicenames.go
  kind: method
  name: PathResolverFunc.Lookup
  signature: func (f PathResolverFunc) Lookup(ctx context.Context, logicalCluster
    string) string
- file: factory/servicenames/servicenames.go
  kind: struct
  name: Resolver
  signature: type Resolver struct { opts Options }
- file: factory/servicenames/servicenames.go
  kind: method
  name: Resolver.Env
  signature: func (d *Resolver) Env(ctx context.Context, target ref.Ref, selfArgs,
    selfEnv string, account *denospec.ServiceAccountRef) map[string]string
- file: factory/servicenames/servicenames.go
  kind: method
  name: Resolver.Name
  signature: func (d *Resolver) Name(ctx context.Context, name, namespace, logicalCluster
    string) string
- file: factory/servicenames/servicenames.go
  kind: method
  name: Resolver.Table
  signature: func (d *Resolver) Table(ctx context.Context) (map[string]string, []string)
- file: factory/servicenames/servicenames.go
  kind: method
  name: Resolver.Tokens
  signature: func (d *Resolver) Tokens(ctx context.Context, workspaces []string, account
    *denospec.ServiceAccountRef) string
- file: factory/servicenames/servicenames.go
  kind: interface
  name: Source
  signature: type Source interface { Pods() []*unstructured.Unstructured }
- file: factory/servicenames/servicenames.go
  kind: method
  name: Source.Pods
  signature: Pods() []*unstructured.Unstructured
- file: factory/servicenames/servicenames.go
  kind: type_alias
  name: SourceFunc
  signature: type SourceFunc func() []*unstructured.Unstructured
- file: factory/servicenames/servicenames.go
  kind: method
  name: SourceFunc.Pods
  signature: func (f SourceFunc) Pods() []*unstructured.Unstructured
requirements:
- codeRefs:
  - function:07389965e6b38ea0b3612bdad786929e
  id: r.advertised-address
  level: MUST
  text: AdvertisedAddress parses the args JSON as a string list and the env JSON as
    a string map, prefers the --port and --hostname flag values over PORT and HOSTNAME,
    returns the empty string when no port is found, and substitutes 127.0.0.1 when
    the host is empty, "0.0.0.0" or "::"; the result is host:port.
- codeRefs:
  - file:factory/servicenames/servicenames.go
  - file:factory/servicenames/servicenames_test.go
  id: r.behaviour-is-tested
  level: MUST
  text: The package behaviour is covered by tests in the same package, including the
    table path lookup, the self name carried before the pod is observed, the shim
    preload entry, and per-workspace token minting.
- codeRefs:
  - method:ad5afdbf8721ab5ce7a03ba3ddcd30e4
  id: r.env-bundle
  level: MUST
  text: Resolver.Env always sets the domain and the target namespace, sets the shim
    key only when Options.Shim is non-empty, adds the JSON table only when the table
    is non-empty, adds the tokens key whenever an account is given, and folds the
    target's own advertised address into the table and the workspace list when that
    address is non-empty.
- codeRefs:
  - method:4278b3fe1612ff604bfe889cf282fc0d
  id: r.name-uses-path
  level: MUST
  text: Resolver.Name routes the logical cluster through the PathResolver when one
    is set and it returns a non-empty path, and otherwise uses the logical cluster
    unchanged, then returns kcp.ServiceFQDN of name, namespace and cluster under the
    configured domain.
- codeRefs:
  - function:b19b27cc115a548b83269276a58c99fc
  - struct:3112320f513f73cfe648eedde3b1fdc4
  - struct:cea2d734e279f4a25d0548390b498f77
  id: r.new-defaults
  level: MUST
  text: New fills Options.Domain with kcp.DefaultServiceDomain and Options.ServiceAccountNamespace
    with "default" when either is empty, and returns a Resolver holding the resulting
    options.
- codeRefs:
  - interface:519ba0a5dc01e2e03c8ae8130b661f15
  - method:3bf4a3c644a5c736a402c5c943469567
  - method:ccfd372e76badab8463bd241875ea2cd
  - type_alias:d0b47bece9defe5890614e1e7fd29a77
  id: r.path-resolver-lookup
  level: MUST
  text: A PathResolver maps a logical cluster to a path, returning the empty string
    when it has no answer, and PathResolverFunc adapts a function with the same signature.
- codeRefs:
  - interface:3f1bd9500073f2f4d05118ef71f4e7fa
  - method:9f382eff81d8cc14c42263ff6cc96655
  - method:ec3d149f59e5f08d064d8abb43c7b47e
  - type_alias:3403f9a23dfea8c2fabcfb5d23057f58
  id: r.source-provides-pods
  level: MUST
  text: A Source returns the observed pods as a slice of unstructured objects, and
    SourceFunc adapts a plain function to that interface by calling it.
- codeRefs:
  - method:af52145174fd9601d169c4b0b89a4a1a
  id: r.table-skips-unlabelled-and-unadvertised
  level: MUST
  text: Resolver.Table returns an empty table and no workspaces when no Source is
    set; it skips pods lacking the kcp.ClusterAnnotation, records each distinct cluster
    once in the returned workspace list in first-seen order, skips pods whose advertised
    address is empty, and maps each remaining pod's resolved name to its advertised
    address.
- codeRefs:
  - method:1d9eaefc92217c5a7cc9fcb25ad6cf14
  id: r.tokens-per-workspace
  level: MUST
  text: Resolver.Tokens mints one service account token per workspace against the
    account's namespace, falling back to Options.ServiceAccountNamespace when the
    account namespace is empty, skips workspaces whose mint fails, and returns the
    result as a JSON object string, or "{}" when marshalling fails. It mints nothing
    when the account is nil or no Minter is set.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:factory/servicenames/servicenames.go` file servicenames.go (factory/servicenames/servicenames.go)
- `file:factory/servicenames/servicenames_test.go` file servicenames_test.go (factory/servicenames/servicenames_test.go)
- `function:07389965e6b38ea0b3612bdad786929e` function AdvertisedAddress (factory/servicenames/servicenames.go)
- `function:b19b27cc115a548b83269276a58c99fc` function New (factory/servicenames/servicenames.go)
- `interface:3f1bd9500073f2f4d05118ef71f4e7fa` interface Source (factory/servicenames/servicenames.go)
- `interface:519ba0a5dc01e2e03c8ae8130b661f15` interface PathResolver (factory/servicenames/servicenames.go)
- `method:1d9eaefc92217c5a7cc9fcb25ad6cf14` method Resolver.Tokens (factory/servicenames/servicenames.go)
- `method:3bf4a3c644a5c736a402c5c943469567` method PathResolver.Lookup (factory/servicenames/servicenames.go)
- `method:4278b3fe1612ff604bfe889cf282fc0d` method Resolver.Name (factory/servicenames/servicenames.go)
- `method:9f382eff81d8cc14c42263ff6cc96655` method SourceFunc.Pods (factory/servicenames/servicenames.go)
- `method:ad5afdbf8721ab5ce7a03ba3ddcd30e4` method Resolver.Env (factory/servicenames/servicenames.go)
- `method:af52145174fd9601d169c4b0b89a4a1a` method Resolver.Table (factory/servicenames/servicenames.go)
- `method:ccfd372e76badab8463bd241875ea2cd` method PathResolverFunc.Lookup (factory/servicenames/servicenames.go)
- `method:ec3d149f59e5f08d064d8abb43c7b47e` method Source.Pods (factory/servicenames/servicenames.go)
- `struct:3112320f513f73cfe648eedde3b1fdc4` struct Resolver (factory/servicenames/servicenames.go)
- `struct:cea2d734e279f4a25d0548390b498f77` struct Options (factory/servicenames/servicenames.go)
- `type_alias:3403f9a23dfea8c2fabcfb5d23057f58` type_alias SourceFunc (factory/servicenames/servicenames.go)
- `type_alias:d0b47bece9defe5890614e1e7fd29a77` type_alias PathResolverFunc (factory/servicenames/servicenames.go)
<!-- SPECD_MANAGED_END -->
