# Context: abc-runner

Repository: `kcp-libs`

This context exists to fix the boundary between the generic workload lifecycle and its concrete implementations. The package names no runtime beyond the field vocabulary of PodRequest, so execrunner, memoryrunner and the examples can implement or consume the same four verbs without the reconciliation, queue and watch layers depending on any one runner. The spec records that contract: the two interfaces, the status and request shapes they exchange, and the State alias both statuses share.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
