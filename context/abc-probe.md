# Context: abc-probe

Repository: `kcp-libs`

This context exists to specify the consecutive-failure probe abstraction: a shared, concurrency-safe place where repeated observations about a named probe key accumulate into a trip/no-trip decision. It is deliberately narrow, holding only counting and threshold normalization, so that callers decide what a key means and what tripping it should do. The behavior encoded here, that a success resets the count, that a new run ID restarts the count, and that keys are independent, is what the tests pin down.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
