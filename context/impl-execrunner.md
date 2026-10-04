# Context: impl-execrunner

Repository: `kcp-libs`

This context exists so the execrunner package has a written specification of the process-execution engine and pod: what the two option structs and their constructors require and default, and what the four lifecycle operations on each runner (Start, Observe, Stop, Probe) must do. It anchors the contract that the runner types implement the generic runner interfaces without knowing about Deno specifics beyond launching a binary, and it records the invariants that the tests in pod_test.go pin down: required directories, path resolution, timeout handling, and non-interference with caller-supplied environment.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
