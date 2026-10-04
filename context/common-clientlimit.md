# Context: common-clientlimit

Repository: `kcp-libs`

This context exists because client-go defaults every client to 5 qps with a burst of 10, which starves a controller that watches and reconciles many objects. The package centralises the fix in one function so each controller construction path (exportwatch, informerwatch, kcpstore) applies the same rate-limit policy without duplicating the defaulting logic, and keeps caller intent: an operator who explicitly sets QPS or Burst on a config is not overridden. It sits in the generic common layer, names no workload runtime, and is usable by any component that builds a rest.Config.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/clientlimit/clientlimit.go` file clientlimit.go (common/clientlimit/clientlimit.go)
- `file:common/clientlimit/clientlimit_test.go` file clientlimit_test.go (common/clientlimit/clientlimit_test.go)
- `function:67bc041de773d39e4803a9795f096d96` function Apply (common/clientlimit/clientlimit.go)
<!-- SPECD_MANAGED_END -->
