# Context: common-kcp

Repository: `kcp-libs`

This context exists so the naming rules that map a KCP logical cluster path onto DNS labels and fully qualified service names stay specified in one place. The rest of the repository, the examples and the service-name factory all depend on the exact string shape produced here, so the reversal order, the root-workspace and empty-segment skipping, the default namespace fallback and the omission of an empty label suffix must hold exactly as the tests assert them.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/kcp/kcp.go` file kcp.go (common/kcp/kcp.go)
- `file:common/kcp/kcp_test.go` file kcp_test.go (common/kcp/kcp_test.go)
- `function:027166c10dcc0d8f2f48420216ff5951` function ServiceFQDN (common/kcp/kcp.go)
- `function:0b14d327c26ec834adfa4a6551d0fc8c` function ServiceLabels (common/kcp/kcp.go)
<!-- SPECD_MANAGED_END -->
