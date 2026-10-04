# Context: impl-memoryrunner

Repository: `kcp-libs`

This context exists to give the rest of kcp-libs a controllable, in-memory fake for the runner contracts (abc/runner), so controllers, reconcile loops and tests can drive pod and engine lifecycles through Start/Observe/Stop/Probe without spawning real processes. It carries the observed tests, which pin the observable behavior: a pod runs for PollsBeforeDone polls before reporting the configured outcome, Stop on a running pod yields StateFailed "stopped", an unknown pod run is an error, and an engine exits after ExitsAfterPolls polls or when stopped with message "stopped". It is deliberately minimal and lives in its own package so the generic layers stay free of any workload-runtime assumption.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/memoryrunner/memoryrunner.go` file memoryrunner.go (impl/memoryrunner/memoryrunner.go)
- `file:impl/memoryrunner/memoryrunner_test.go` file memoryrunner_test.go (impl/memoryrunner/memoryrunner_test.go)
- `function:e2306bd41881382147eaf8b3069895ee` function NewEngine (impl/memoryrunner/memoryrunner.go)
- `function:e28103a134b3e9ac6daaf0bb72e41005` function NewPod (impl/memoryrunner/memoryrunner.go)
- `method:098390ace6e69e0022deef1529cb656b` method Engine.Observe (impl/memoryrunner/memoryrunner.go)
- `method:48226f7f64a37307caf8965bccf68855` method Engine.Start (impl/memoryrunner/memoryrunner.go)
- `method:49f3a507847759e15ebe12029f4d7ab6` method Engine.Probe (impl/memoryrunner/memoryrunner.go)
- `method:7a9767b282a97214edbe56a9c367cbf9` method Pod.Probe (impl/memoryrunner/memoryrunner.go)
- `method:806b7d8c875ce7fdb1779fc6b3eb84f8` method Pod.Start (impl/memoryrunner/memoryrunner.go)
- `method:d13dc67d5b49f32d887748200e8e6880` method Pod.Observe (impl/memoryrunner/memoryrunner.go)
- `method:e9fd6d2481d5bc0aad2f394f8ba186eb` method Pod.Stop (impl/memoryrunner/memoryrunner.go)
- `method:ee1b796b776a0010bd5381e7774de163` method Engine.Stop (impl/memoryrunner/memoryrunner.go)
- `struct:9d329f0ad839a867f04ada3209850157` struct Engine (impl/memoryrunner/memoryrunner.go)
- `struct:ac6c667397cefb3293ada1d745b0df58` struct PodOptions (impl/memoryrunner/memoryrunner.go)
- `struct:ee18afe6dcdfa47419c426a3dd83071d` struct EngineOptions (impl/memoryrunner/memoryrunner.go)
- `struct:f1cb1d1b9762a5263f685d1b1a57baa4` struct Pod (impl/memoryrunner/memoryrunner.go)
<!-- SPECD_MANAGED_END -->
