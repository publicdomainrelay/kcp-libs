# Context: factory-servicenames

Repository: `kcp-libs`

This context exists so that workloads running under kcp can discover each other by name without a hand-written DNS configuration. It converts the pods a cluster controller already observes into a service table (name to advertised host:port), the workspace list worth addressing, and the service account tokens needed to reach them, then hands that whole bundle to a workload as environment variables. It also defines the seams (Source, PathResolver, TokenMinter) that keep the package free of any particular client, so a caller supplies real store-backed pods and paths while tests supply plain functions.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
