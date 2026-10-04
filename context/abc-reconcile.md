# Context: abc-reconcile

Repository: `kcp-libs`

This context exists to separate the decision half of a controller from the transport and runtime half. A Reconciler is pure: given the observed object it returns what it wants to happen, with no client, no workqueue and no dependency on any workload runtime. A Handler wraps the other direction, owning retries and terminality so the reconcile loop can be scheduled; Policy centralises the backoff arithmetic instead of scattering it through call sites, and Bridge is the piece that actually reads, decides and applies against a cluster. Result and Patch exist so the decided outcome can be accumulated, interrogated and emitted as a patch without the reconciler having to know how the write is performed. The package names no workload runtime so the generic layers of the library stay usable without the deno packages.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
