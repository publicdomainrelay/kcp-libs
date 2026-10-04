# Context: internal-livekcp-livetest

Repository: `kcp-libs`

This context exists so that live tests, which need real kcp and kubectl binaries, can degrade gracefully on machines and CI runners that do not have those tools. Instead of each test repeating PATH probing and environment checks, they call livetest.Require(t) and inherit one consistent policy: skip by default, and fail loudly only when the operator has explicitly demanded the live tier by setting KCP_LIBS_REQUIRE_LIVE=1. The helper is the single switch that turns an entire test tier on and off.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: internal/livekcp/livetest/livetest.go
  kind: function
  name: Require
  signature: func Require(t *testing.T)
requirements:
- codeRefs:
  - file:internal/livekcp/livetest/livetest.go
  - function:42a6190b9b8e226427c4db2ace96f2cf
  id: r.fail-when-required-and-tools-missing
  level: MUST
  text: When KCP_LIBS_REQUIRE_LIVE equals "1" and livekcp.Missing() reports at least
    one absent binary, Require must fail the test immediately with t.Fatalf naming
    the missing binaries, so a demanded live run never silently passes.
- codeRefs:
  - file:internal/livekcp/livetest/livetest.go
  - function:42a6190b9b8e226427c4db2ace96f2cf
  id: r.mark-helper
  level: SHOULD
  text: Require should call t.Helper() so failures and skips are attributed to the
    calling test rather than to the helper.
- codeRefs:
  - file:internal/livekcp/livetest/livetest.go
  - function:42a6190b9b8e226427c4db2ace96f2cf
  id: r.return-when-required-and-tools-present
  level: MUST
  text: When KCP_LIBS_REQUIRE_LIVE equals "1" and livekcp.Missing() reports nothing
    missing, Require must return without skipping so the calling live test runs.
- codeRefs:
  - file:internal/livekcp/livetest/livetest.go
  - function:42a6190b9b8e226427c4db2ace96f2cf
  id: r.skip-naming-missing-tools
  level: MUST
  text: When the live tier is not required and livekcp.Missing() reports absent binaries,
    Require must skip the test with t.Skipf and a message that names those binaries.
- codeRefs:
  - file:internal/livekcp/livetest/livetest.go
  - function:42a6190b9b8e226427c4db2ace96f2cf
  id: r.skip-when-not-required
  level: MUST
  text: When the live tier is not required and no binaries are missing, Require must
    still skip the test with t.Skipf, telling the caller to set KCP_LIBS_REQUIRE_LIVE=1
    to run the live tier.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:internal/livekcp/livetest/livetest.go` file livetest.go (internal/livekcp/livetest/livetest.go)
- `function:42a6190b9b8e226427c4db2ace96f2cf` function Require (internal/livekcp/livetest/livetest.go)
<!-- SPECD_MANAGED_END -->
