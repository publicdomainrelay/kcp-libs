# Context: common-denocomputer

Repository: `kcp-libs`

This context exists to keep one copy of the deno.computer names and phases that both this generic library and the deno-kcp consumer agree on. Rather than letting the library invent its own group, finalizers, labels, conditions or phase strings, the constants live here and a test reads the consumer's own source as the source of truth and fails on any drift. The phase predicates sit beside the Phase type so callers in the reconcile paths can ask whether a run, pod, job, policy workflow or policy engine has finished without re-encoding which phases count as terminal in every call site.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
