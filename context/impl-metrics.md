# Context: impl-metrics

Repository: `kcp-libs`

This context exists so that every component in the repository can expose Prometheus metrics through one small, uniform surface instead of touching the prometheus client directly. It gives each component an isolated registry plus a naming prefix, so two components in one process cannot collide on a metric name, and it makes the common collectors idempotent by name so repeated wiring is safe. It also supplies the two export paths a process needs: Render for in-process encoding and Listen for an HTTP scrape endpoint, with Server.Address and Server.Close for lifecycle control and tests that bind port zero.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
