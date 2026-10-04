# Context: impl-execrunner

Repository: `kcp-libs`

This context exists so the execrunner package has a written specification of the process-execution engine and pod: what the two option structs and their constructors require and default, and what the four lifecycle operations on each runner (Start, Observe, Stop, Probe) must do. It anchors the contract that the runner types implement the generic runner interfaces without knowing about Deno specifics beyond launching a binary, and it records the invariants that the tests in pod_test.go pin down: required directories, path resolution, timeout handling, and non-interference with caller-supplied environment.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/execrunner/engine.go` file engine.go (impl/execrunner/engine.go)
- `file:impl/execrunner/pod.go` file pod.go (impl/execrunner/pod.go)
- `file:impl/execrunner/pod_test.go` file pod_test.go (impl/execrunner/pod_test.go)
- `file:impl/execrunner/supervisor.go` file supervisor.go (impl/execrunner/supervisor.go)
- `function:00b074a7109347b2daa577f4bdc75467` function NewPod (impl/execrunner/pod.go)
- `function:4c978aedd11952f3ee6365eb5898c321` function NewEngine (impl/execrunner/engine.go)
- `method:0cfcac9db3f96294b49ce881bee14d93` method Pod.Observe (impl/execrunner/pod.go)
- `method:2beee7604d9afe94ca0951259ecada45` method Engine.Start (impl/execrunner/engine.go)
- `method:42a8aeb02623810d653a714732142368` method Engine.Stop (impl/execrunner/engine.go)
- `method:53d6600d1f29a1f604f9fe12de4fe3df` method Engine.Probe (impl/execrunner/engine.go)
- `method:55c0a5d0dd6cf0808dd2503822444b65` method Pod.Stop (impl/execrunner/pod.go)
- `method:e4363e67372d0067b655391e57db53ed` method Pod.Start (impl/execrunner/pod.go)
- `method:e837684e27d5363ecf736c655e7e5b24` method Engine.Observe (impl/execrunner/engine.go)
- `method:ef6ab966433b9776496a384adb8cf1e1` method Pod.Probe (impl/execrunner/pod.go)
- `struct:5a9b7da405a97b69724f6f93d52cf4ab` struct Engine (impl/execrunner/engine.go)
- `struct:6aef5701957707ea4bef629407cabc37` struct EngineOptions (impl/execrunner/engine.go)
- `struct:aa3532720785693eda3016d5bd68272c` struct Pod (impl/execrunner/pod.go)
- `struct:c617bd87489adfc74952ca77d12f02ae` struct PodOptions (impl/execrunner/pod.go)
<!-- SPECD_MANAGED_END -->
