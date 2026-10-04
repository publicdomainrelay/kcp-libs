# Context: impl-kcpstore

Repository: `kcp-libs`

This context exists so that consumers (controllers, admission, DNS, servicenames) have one place that knows how to reach kcp: URL shape, per-cluster client reuse, rate limiting, JSON decoding, status and finalizer patching, token minting, and workspace path resolution. The path cache is the newest part and it is deliberately conservative — it exists because a workspace path that failed to resolve must not be frozen into a wrong name, so only an answer that cannot change by asking again is stored.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
