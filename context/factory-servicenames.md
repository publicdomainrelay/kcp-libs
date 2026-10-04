# Context: factory-servicenames

Repository: `kcp-libs`

This context exists so that workloads running under kcp can discover each other by name without a hand-written DNS configuration. It converts the pods a cluster controller already observes into a service table (name to advertised host:port), the workspace list worth addressing, and the service account tokens needed to reach them, then hands that whole bundle to a workload as environment variables. It also defines the seams (Source, PathResolver, TokenMinter, the workspace source) that keep the package free of any particular client, so a caller supplies real store-backed pods, paths and workspaces while tests supply plain functions. The token map and the service names agree on one cluster naming on purpose: a workload discovers a peer by parsing the peer's name back to its cluster and looking the token up by it, so a token keyed any other way is a peer it can never reach.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
