# Context: common-ttl

Repository: `kcp-libs`

This context exists so callers do not each re-derive TTL semantics from raw pointers. Kubernetes-style APIs express TTLs as optional seconds on a timestamp, where nil means unset and negative is meaningless; these helpers centralize that normalization so every consumer agrees on when something is scheduled for deletion, and when nothing should be scheduled at all.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
