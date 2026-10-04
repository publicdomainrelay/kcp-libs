# Context: factory-controller

Repository: `kcp-libs`

This context is the controller runtime of kcp-libs: the generic work-queue driver that turns watched kcp resources into reconcile invocations. It exists so the deno workload packages and downstream consumers can share one tested scheduling core — bounded-concurrency workers, rate limited retries, optimistic-concurrency conflict handling, informer-event freshness tracking and prometheus metrics — instead of each rebuilding it. It names no workload runtime; the deno vocabulary lives in the consumers, and the controller only knows kinds, refs and a reconcile.Handler.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
