# Context: factory-admission

Repository: `kcp-libs`

This context exists to describe the admission gate that sits between a queue and the workloads it schedules. It defines the Source port through which the gate reads cluster state (parent, siblings, capacity) without knowing any workload runtime, and the Admitter that turns those reads plus a lease table into a queue.Admission decision and into wake notifications for runs that have become eligible. It is the piece that enforces per-parent concurrency policy, so callers can ask whether a specific run may proceed and can be told which runs to wake once capacity frees up.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
