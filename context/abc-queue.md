# Context: abc-queue

Repository: `kcp-libs`

This context exists to specify the admission/lease contract of the abc queue: what a Run must look like to be planned, how Policy defaults resolve, how a Capacity, Blocker and reserved count combine into a per-run Admission, and how held leases are scoped to a parent, counted against observed lifecycle phase, expired by TTL and released. It is the durable description of the queue's pure decision functions and its lease bookkeeping, so consumers such as the deno process runner and the admission factory can rely on the same semantics without re-deriving them from the source.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
