# Context: examples-controller

Repository: `kcp-libs`

This context exists to demonstrate, end to end and against a real kcp instance, that the generic controller layer works: a store backed by kcp, an export watch that discovers the virtual workspace URL, an informer-based controller with indexers and a bounded reconcile policy, and metrics. It is an executable example and a test fixture (main_test.go), not library code, so it composes the generic packages (kcpstore, exportwatch, cache, informerwatch, controller, metrics) rather than defining new abstractions.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/controller/main.go` file main.go (examples/controller/main.go)
- `file:examples/controller/main_test.go` file main_test.go (examples/controller/main_test.go)
- `function:23c706f296c8210aa4211fe5ad7815cf` function Run (examples/controller/main.go)
<!-- SPECD_MANAGED_END -->
