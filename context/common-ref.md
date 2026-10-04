# Context: common-ref

Repository: `kcp-libs`

This context exists so that every layer of kcp-libs agrees on the identity and the transport address of a KCP object. A Ref is the identity used as a map key, a cache index key and a workqueue key, so its string form must be stable and shared; BaseHost and ClusterURL exist so that callers holding either a bare host or a host that already carries a /clusters/<logicalCluster> path can still build a correct per-cluster URL. Keeping these helpers in one tiny package prevents the key format and the KCP API path prefix from being re-implemented, and diverging, in each consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
