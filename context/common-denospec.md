# Context: common-denospec

Repository: `kcp-libs`

This context exists so that the deno workload vocabulary lives in one Go package that both the KCP libraries and the Deno-side consumer agree on. The structs are the wire contract; the contract test fails if the Go shapes drift from what the consumer declares. Validate and Args exist so that a permission set declared as data becomes a checked, runnable Deno command line rather than an ad-hoc string built at each call site, and EffectiveRestartPolicy and ProbeCommand pin down the remaining runtime decisions around restarts and probes.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/denospec/denospec.go` file denospec.go (common/denospec/denospec.go)
- `file:common/denospec/denospec_contract_test.go` file denospec_contract_test.go (common/denospec/denospec_contract_test.go)
- `file:common/denospec/denospec_test.go` file denospec_test.go (common/denospec/denospec_test.go)
- `function:10ed3675a84f3ace014b3a75a3e36f7c` function ProbeCommand (common/denospec/denospec.go)
- `function:8c740b4bfa20b8c5c3abcbedabd84404` function EffectiveRestartPolicy (common/denospec/denospec.go)
- `function:8d4402900d3462a266b3860020572db3` function Args (common/denospec/denospec.go)
- `function:961db49fe00a81882f3e64de7ff7262d` function Validate (common/denospec/denospec.go)
- `struct:37cfecc71eb8d4169a8fdffff71f2eb3` struct Permissions (common/denospec/denospec.go)
- `struct:81393b3c6451cfbfd692bcbb851b30a9` struct ServiceAccountRef (common/denospec/denospec.go)
- `struct:bbb7094c15f71e3d40b7bd6ef4ffad5e` struct Permission (common/denospec/denospec.go)
- `struct:e63b3474fb8e8342ee8ddcbf9aeff650` struct ExecProbe (common/denospec/denospec.go)
- `struct:ee22e42a28e08358a06298c6048e62e5` struct PodTemplate (common/denospec/denospec.go)
- `type_alias:c9ed92b47f68b4cc5b0bb51223210dbb` type_alias RestartPolicy (common/denospec/denospec.go)
<!-- SPECD_MANAGED_END -->
