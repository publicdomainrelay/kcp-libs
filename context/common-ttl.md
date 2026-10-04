# Context: common-ttl

Repository: `kcp-libs`

This context exists so callers do not each re-derive TTL semantics from raw pointers. Kubernetes-style APIs express TTLs as optional seconds on a timestamp, where nil means unset and negative is meaningless; these helpers centralize that normalization so every consumer agrees on when something is scheduled for deletion, and when nothing should be scheduled at all.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/ttl/ttl.go` file ttl.go (common/ttl/ttl.go)
- `file:common/ttl/ttl_test.go` file ttl_test.go (common/ttl/ttl_test.go)
- `function:503ea477e92bcc1bc7f9bdba99922323` function Expiry (common/ttl/ttl.go)
- `function:9fdc74d33c3a88840f75732e3c0170bb` function Deadline (common/ttl/ttl.go)
- `function:a0c1200f8b55a18bfa918f1f0cee7615` function Expired (common/ttl/ttl.go)
- `function:d687afb3744f79be091b78a990794a9d` function Effective (common/ttl/ttl.go)
<!-- SPECD_MANAGED_END -->
