# Context: common-denospec

Repository: `kcp-libs`

This context exists so that the deno workload vocabulary lives in one Go package that both the KCP libraries and the Deno-side consumer agree on. The structs are the wire contract; the contract test fails if the Go shapes drift from what the consumer declares. Validate and Args exist so that a permission set declared as data becomes a checked, runnable Deno command line rather than an ad-hoc string built at each call site, and EffectiveRestartPolicy and ProbeCommand pin down the remaining runtime decisions around restarts and probes.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
