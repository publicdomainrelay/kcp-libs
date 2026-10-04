# Context: examples-admission

Repository: `kcp-libs`

This context exists so that the admission example can be described and regenerated as a compile-checked demonstration of the admission Source interface. It documents the exact contract an implementer must satisfy: how a parent is discovered from a run (a label lookup that may legitimately find nothing), how child runs are enumerated and filtered by that same label, and how capacity is derived from the parent batch, including the not-found case which must surface as a Blocker rather than an error. Keeping the spec anchored to these methods lets changes to the queue capacity and blocker types be validated against a real consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
