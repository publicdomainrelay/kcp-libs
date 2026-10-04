# Context: impl-policyclient

Repository: `kcp-libs`

This context exists to give the rest of the repository one place that speaks the policy engine's HTTP protocol, so callers depend on the abc/policy.Client interface and never build requests themselves. It exists because policy runs are submitted as JSON workflows to an engine endpoint and their verdicts come back as task status, and both halves need the same validation, error wrapping and output flattening.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
