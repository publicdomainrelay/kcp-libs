# Context: impl-exportwatch

Repository: `kcp-libs`

This context exists so callers can wait until a provider's exports are actually reachable before proceeding: it isolates the endpoint-slice discovery, projection and readiness logic from any workload runtime, gives a single place where the provider workspace is appended to the API host path, and exposes both a blocking wait (Await) and a watch-driven wait (WatchEndpointSlices) over the same Options.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/exportwatch/exportwatch.go` file exportwatch.go (impl/exportwatch/exportwatch.go)
- `function:0e22ac2739d17dd3caab2abd75c390e0` function Discover (impl/exportwatch/exportwatch.go)
- `function:4855b61ac47e7bbc0bb211b8661e9d59` function FromList (impl/exportwatch/exportwatch.go)
- `function:48d584db657cf9d26e0c376b34aa6773` function Await (impl/exportwatch/exportwatch.go)
- `function:96edd4eb5a20bcded694ab307fa5f365` function Paths (impl/exportwatch/exportwatch.go)
- `function:b185d3ac6de1d5bfcfd671f7afeca0c2` function WatchEndpointSlices (impl/exportwatch/exportwatch.go)
- `function:efb61af66b10e227581818288975a28e` function Client (impl/exportwatch/exportwatch.go)
- `method:631983bf08fd7d3133061275ab1f910d` method Endpoints.Ready (impl/exportwatch/exportwatch.go)
- `struct:99a8a7585a6b5d30375c471a8fb38c71` struct Options (impl/exportwatch/exportwatch.go)
- `type_alias:94252f87faff3176628078b9500af709` type_alias Endpoints (impl/exportwatch/exportwatch.go)
<!-- SPECD_MANAGED_END -->
