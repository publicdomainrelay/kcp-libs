# Context: abc-store

Repository: `kcp-libs`

This context exists to fix the storage contract that the rest of kcp-libs codes against. Controllers, caches and reconcilers need to read and write cluster objects and mint service account tokens, but they must not depend on a concrete Kubernetes client or on a specific workload runtime. abc-store supplies that seam as small generic interfaces plus one pure comparison helper. Unchanged exists because status patches should only be sent when the object actually changed on the wire, and that comparison must follow JSON encoding rules rather than Go struct identity.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
