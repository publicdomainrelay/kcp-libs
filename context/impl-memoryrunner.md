# Context: impl-memoryrunner

Repository: `kcp-libs`

This context exists to give the rest of kcp-libs a controllable, in-memory fake for the runner contracts (abc/runner), so controllers, reconcile loops and tests can drive pod and engine lifecycles through Start/Observe/Stop/Probe without spawning real processes. It carries the observed tests, which pin the observable behavior: a pod runs for PollsBeforeDone polls before reporting the configured outcome, Stop on a running pod yields StateFailed "stopped", an unknown pod run is an error, and an engine exits after ExitsAfterPolls polls or when stopped with message "stopped". It is deliberately minimal and lives in its own package so the generic layers stay free of any workload-runtime assumption.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
