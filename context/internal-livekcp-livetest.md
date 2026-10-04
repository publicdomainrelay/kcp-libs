# Context: internal-livekcp-livetest

Repository: `kcp-libs`

This context exists so that live tests, which need real kcp and kubectl binaries, can degrade gracefully on machines and CI runners that do not have those tools. Instead of each test repeating PATH probing and environment checks, they call livetest.Require(t) and inherit one consistent policy: skip by default, and fail loudly only when the operator has explicitly demanded the live tier by setting KCP_LIBS_REQUIRE_LIVE=1. The helper is the single switch that turns an entire test tier on and off.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:internal/livekcp/livetest/livetest.go` file livetest.go (internal/livekcp/livetest/livetest.go)
- `function:42a6190b9b8e226427c4db2ace96f2cf` function Require (internal/livekcp/livetest/livetest.go)
<!-- SPECD_MANAGED_END -->
