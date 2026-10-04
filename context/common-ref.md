# Context: common-ref

Repository: `kcp-libs`

This context exists so that every layer of kcp-libs agrees on the identity and the transport address of a KCP object. A Ref is the identity used as a map key, a cache index key and a workqueue key, so its string form must be stable and shared; BaseHost and ClusterURL exist so that callers holding either a bare host or a host that already carries a /clusters/<logicalCluster> path can still build a correct per-cluster URL. Keeping these helpers in one tiny package prevents the key format and the KCP API path prefix from being re-implemented, and diverging, in each consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/ref/ref.go` file ref.go (common/ref/ref.go)
- `file:common/ref/ref_test.go` file ref_test.go (common/ref/ref_test.go)
- `function:4c5cf31e71ec19179fbbb42d890195d5` function ClusterURL (common/ref/ref.go)
- `function:757cbe1304d5ad40e0627224acf0be64` function BaseHost (common/ref/ref.go)
- `function:7999d84f451c59ca6a03577995281764` function New (common/ref/ref.go)
- `function:a77844fa6f24528303d849e5ed956c3c` function Key (common/ref/ref.go)
- `method:2948085699170c0e8f98592fb74fedb5` method Ref.WithNamespace (common/ref/ref.go)
- `method:4e7c59fd75778a303dccf1476257f32b` method Ref.Key (common/ref/ref.go)
- `method:a10e7cf01872c2cd075539e6dba8e9f4` method Ref.WithResourceVersion (common/ref/ref.go)
- `struct:d0d183e54e296ffe32b06ffd08b347ca` struct Ref (common/ref/ref.go)
<!-- SPECD_MANAGED_END -->
