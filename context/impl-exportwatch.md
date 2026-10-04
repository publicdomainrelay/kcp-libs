# Context: impl-exportwatch

Repository: `kcp-libs`

This context exists so callers can wait until a provider's exports are actually reachable before proceeding: it isolates the endpoint-slice discovery, projection and readiness logic from any workload runtime, gives a single place where the provider workspace is appended to the API host path, and exposes both a blocking wait (Await) and a watch-driven wait (WatchEndpointSlices) over the same Options.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
