# Context: impl-informerwatch

Repository: `kcp-libs`

This context exists so the kcp-libs dynamic watching layer is described as it is written: a runtime-neutral informer bootstrap that turns kcp object events into enqueue calls and optional reactor callbacks, with the indexer, cache set and reactor hooks injected by the caller. It names no workload runtime and is reusable without the deno packages.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/informerwatch/informerwatch.go` file informerwatch.go (impl/informerwatch/informerwatch.go)
- `file:impl/informerwatch/informerwatch_test.go` file informerwatch_test.go (impl/informerwatch/informerwatch_test.go)
- `function:e7b0c39be09f14cf5d6e760c26e08195` function Run (impl/informerwatch/informerwatch.go)
- `interface:69ae3bee91c7100683df536ca4bee1f1` interface Reactor (impl/informerwatch/informerwatch.go)
- `method:19895db5d0e36b28d57bf850b23b04ad` method Reactor.Added (impl/informerwatch/informerwatch.go)
- `method:965fbef5d03cd976d93d50b7588840e0` method Reactor.Updated (impl/informerwatch/informerwatch.go)
- `struct:591bf69eb781348c20c82bf16173a503` struct Options (impl/informerwatch/informerwatch.go)
- `struct:f606eb0070d5227633ed9759d17ac235` struct Resource (impl/informerwatch/informerwatch.go)
- `struct:ff54a13981e71b687c7ac0015b595ed9` struct Source (impl/informerwatch/informerwatch.go)
- `type_alias:922cbb149aec17f0c3fd708c19ebcc25` type_alias Enqueue (impl/informerwatch/informerwatch.go)
<!-- SPECD_MANAGED_END -->
