# Context: factory-servicenames

Repository: `kcp-libs`

This context exists so that workloads running under kcp can discover each other by name without a hand-written DNS configuration. It converts the pods a cluster controller already observes into a service table (name to advertised host:port), the workspace list worth addressing, and the service account tokens needed to reach them, then hands that whole bundle to a workload as environment variables. It also defines the seams (Source, PathResolver, TokenMinter, the workspace source) that keep the package free of any particular client, so a caller supplies real store-backed pods, paths and workspaces while tests supply plain functions. The token map and the service names agree on one cluster naming on purpose: a workload discovers a peer by parsing the peer's name back to its cluster and looking the token up by it, so a token keyed any other way is a peer it can never reach.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:factory/servicenames/servicenames.go` file servicenames.go (factory/servicenames/servicenames.go)
- `file:factory/servicenames/servicenames_test.go` file servicenames_test.go (factory/servicenames/servicenames_test.go)
- `function:37ab5725bd1c20dbfd19a629daa2966c` function New (factory/servicenames/servicenames.go)
- `function:8a334a3b9191ecba1125496302619eef` function AdvertisedAddress (factory/servicenames/servicenames.go)
- `interface:3f1bd9500073f2f4d05118ef71f4e7fa` interface Source (factory/servicenames/servicenames.go)
- `interface:4ebaec1f74109d6dcdd5736cd559e3a4` interface Workspaces (factory/servicenames/servicenames.go)
- `interface:c70a17259fae28f88b3e1922fd769641` interface PathResolver (factory/servicenames/servicenames.go)
- `method:21dae96a0de0787a3171d9b273266bc6` method Resolver.Tokens (factory/servicenames/servicenames.go)
- `method:7f75929e6209618a7c76aae36b0b0ced` method WorkspacesFunc.Clusters (factory/servicenames/servicenames.go)
- `method:9f382eff81d8cc14c42263ff6cc96655` method SourceFunc.Pods (factory/servicenames/servicenames.go)
- `method:9fac36ee8b8d24e43bbc4732929d149c` method PathResolver.Lookup (factory/servicenames/servicenames.go)
- `method:bfae8f5500472858636c6d1c013433aa` method Resolver.Env (factory/servicenames/servicenames.go)
- `method:cae3fb88acdade71c05e6f034630da80` method PathResolverFunc.Lookup (factory/servicenames/servicenames.go)
- `method:cf182c7f9ed04c1919dc7b4d08b33cad` method Workspaces.Clusters (factory/servicenames/servicenames.go)
- `method:ec3d149f59e5f08d064d8abb43c7b47e` method Source.Pods (factory/servicenames/servicenames.go)
- `method:edee24a98f35bb7511a0389d67d2ebbf` method Resolver.Table (factory/servicenames/servicenames.go)
- `method:fde6488c215fafdc9b615405664720f8` method Resolver.Name (factory/servicenames/servicenames.go)
- `struct:6a9fe9f2761d8754abd4ba3af073a35f` struct Options (factory/servicenames/servicenames.go)
- `struct:dd81ce362a0aa33a9ebf018c6e22efaa` struct Resolver (factory/servicenames/servicenames.go)
- `type_alias:18970bc4f4801584a91a93a1396a18a1` type_alias PathResolverFunc (factory/servicenames/servicenames.go)
- `type_alias:3403f9a23dfea8c2fabcfb5d23057f58` type_alias SourceFunc (factory/servicenames/servicenames.go)
- `type_alias:a6d9b9d4187d6bb1511f3a5ec95fa053` type_alias WorkspacesFunc (factory/servicenames/servicenames.go)
<!-- SPECD_MANAGED_END -->
