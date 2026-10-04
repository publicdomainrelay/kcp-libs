# Context: common-expiringmap

Repository: `kcp-libs`

This context exists as the shared expiry primitive for the wider kcp-libs project: callers that need a bounded-lifetime cache (for example the job allocator's lookup tables) get one place that owns the TTL comparison, the lazy delete on read, and the locking discipline, rather than each site re-deriving `at.Sub(entry.At) >= ttl` and remembering to hold a lock while doing it. Passing `at` in rather than calling time.Now keeps the expiry behaviour deterministic and testable, which is exactly how the accompanying test file drives it.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
