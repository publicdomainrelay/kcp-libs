# Context: examples-dns

Repository: `kcp-libs`

This context exists to demonstrate and verify the DNS and service-name resolution behaviour of kcp-libs end to end. It is an executable example rather than library code: it shows a consumer of `servicenames` how to assemble a resolver over several logical clusters, how tokens for the kcp server are minted and injected into a pod environment, and how the DNS shim and readiness probe assets are placed on disk so a workload can resolve sibling services. The test in `main_test.go` pins the observable output of that flow, so a change in table contents, token minting, asset placement, advertised address, or label rendering fails the example.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
