# Context: abc-runner

Repository: `kcp-libs`

This context exists to fix the boundary between the generic workload lifecycle and its concrete implementations. The package names no runtime beyond the field vocabulary of PodRequest, so execrunner, memoryrunner and the examples can implement or consume the same four verbs without the reconciliation, queue and watch layers depending on any one runner. The spec records that contract: the two interfaces, the status and request shapes they exchange, and the State alias both statuses share.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/runner/runner.go` file runner.go (abc/runner/runner.go)
- `interface:710591a1c2c64a81c34bf809d899679a` interface PodRunner (abc/runner/runner.go)
- `interface:cee292400f7e239797b1b10cd6e5e2f3` interface EngineRunner (abc/runner/runner.go)
- `method:0dd0d42d1a37091f0638d086e94ed5c6` method PodRunner.Observe (abc/runner/runner.go)
- `method:15f9ba905e255f2171396063a3091a74` method PodRunner.Stop (abc/runner/runner.go)
- `method:193f9210bd1837651faf136898a881d1` method EngineRunner.Observe (abc/runner/runner.go)
- `method:387e9c4ad715a0e08a6e5e5208da6a72` method EngineRunner.Probe (abc/runner/runner.go)
- `method:405b17e18e48a4f4ac2253ec9ce449dd` method PodRunner.Probe (abc/runner/runner.go)
- `method:9b3f09bbc26d631a7a08a5c4713097bc` method EngineRunner.Start (abc/runner/runner.go)
- `method:fdb656c1b22baaa050976fef1ccc2950` method PodRunner.Start (abc/runner/runner.go)
- `method:febb512db7e00e28129ca183c51968b0` method EngineRunner.Stop (abc/runner/runner.go)
- `struct:46cec8074c83b78c8fabcd0fc4957e32` struct PodRequest (abc/runner/runner.go)
- `struct:5dc6f04f8f1d6725396ad778f57753ed` struct PodStatus (abc/runner/runner.go)
- `struct:b8dec15e30496deaf9bd895e8e7218d8` struct EngineStatus (abc/runner/runner.go)
- `struct:bd7270a885b505237f5cd4d9038da86d` struct EngineRequest (abc/runner/runner.go)
- `type_alias:5468a2c306be0c500554229bdda5aadd` type_alias State (abc/runner/runner.go)
<!-- SPECD_MANAGED_END -->
