# Context: internal-livekcp

Repository: `kcp-libs`

This context exists so tests can run against a real kcp control plane instead of mocks, with a stable Go API for lifecycle, inspection, and object seeding. It separates tool discovery (Resolve, Missing) from cluster lifecycle (Start, Cluster.Stop) and from object manipulation (NewObject, Seed, WaitFor), so the live tier is opt-in and its prerequisites are checkable before a test starts.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
