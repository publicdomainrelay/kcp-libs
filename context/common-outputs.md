# Context: common-outputs

Repository: `kcp-libs`

This context exists to fix the contract for converting arbitrary JSON-ish values into the string map that kcp surfaces as outputs. The two functions are the single place where that conversion rule lives, so the exec runner and the policy client cannot drift apart on how a number, boolean, or list becomes a string. The spec records the observable contract: nil in gives nil out, strings pass through unchanged, non-strings become their JSON encoding, and a value JSON cannot encode still yields a usable string rather than an error.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/outputs/outputs.go` file outputs.go (common/outputs/outputs.go)
- `file:common/outputs/outputs_test.go` file outputs_test.go (common/outputs/outputs_test.go)
- `function:1f7915d8cd9aa1f3759d833f5af90128` function StringifyValue (common/outputs/outputs.go)
- `function:d16f38628fc00d680c6f54309865e821` function Stringify (common/outputs/outputs.go)
<!-- SPECD_MANAGED_END -->
