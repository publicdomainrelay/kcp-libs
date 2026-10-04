# Context: examples-workloads

Repository: `kcp-libs`

This context exists to prove the workload packages compose in one place and to serve as living documentation for callers. It is an example, not a library: it is a package main whose only exported symbol is Run, deliberately taking a context and an io.Writer so the same code path can be driven by a test that captures output instead of the terminal. The accompanying test asserts on the printed lines, which makes the example double as an integration check that the deno permission vocabulary, the exec runner's argv layout, probe thresholds, ttl arithmetic, job allocation bookkeeping and the memory runner all keep behaving as documented.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/workloads/main.go` file main.go (examples/workloads/main.go)
- `file:examples/workloads/main_test.go` file main_test.go (examples/workloads/main_test.go)
- `function:4f9e3c5199d0b48e53740bacffb4b9ac` function Run (examples/workloads/main.go)
<!-- SPECD_MANAGED_END -->
