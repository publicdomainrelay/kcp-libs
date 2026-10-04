# Context: impl-assets

Repository: `kcp-libs`

This context exists so callers can hand a declarative set of asset files to one object and get back a stable name-to-path map, without caring when or how often the bytes hit disk. It separates one-shot materialisation, which is idempotent and retried after failure, from explicit rewriting, which always writes, and it gives the DNS workload a ready-made set for its shim and probe scripts.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
