# Context: examples-workloads

Repository: `kcp-libs`

This context exists to prove the workload packages compose in one place and to serve as living documentation for callers. It is an example, not a library: it is a package main whose only exported symbol is Run, deliberately taking a context and an io.Writer so the same code path can be driven by a test that captures output instead of the terminal. The accompanying test asserts on the printed lines, which makes the example double as an integration check that the deno permission vocabulary, the exec runner's argv layout, probe thresholds, ttl arithmetic, job allocation bookkeeping and the memory runner all keep behaving as documented.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: examples/workloads/main.go
  kind: function
  name: Run
  signature: func Run(ctx context.Context, out io.Writer) error
requirements:
- codeRefs:
  - file:examples/workloads/main.go
  id: r.exec-pod-lifecycle
  level: MUST
  text: Run starts a workload through the execrunner pod, asserts it satisfies runner.PodRunner,
    waits until the run leaves the running state, and reports the terminal state and
    the answer output.
- codeRefs:
  - file:examples/workloads/main.go
  id: r.failure-and-probe-paths-shown
  level: MUST
  text: Run must also exercise a failing run that reports a non-zero exit code, a
    readiness probe that passes, and a liveness tracker that requests a restart once
    the failure threshold is reached.
- codeRefs:
  - file:examples/workloads/main.go
  id: r.memory-runner-parity
  level: SHOULD
  text: Run replays the greeting script through the in-memory runner and shows it
    produces the same answer without starting a process.
- codeRefs:
  - file:examples/workloads/main.go
  id: r.permissions-become-argv
  level: MUST
  text: Run converts a denospec.Permissions value with a net allow-list and denied
    env into argument strings via denospec.Args, and prints them.
- codeRefs:
  - file:examples/workloads/main.go
  - function:4f9e3c5199d0b48e53740bacffb4b9ac
  id: r.run-is-the-entry-point
  level: MUST
  text: The example exposes Run(ctx context.Context, out io.Writer) error; it takes
    the context and the output writer as parameters and returns an error rather than
    terminating the process, so a caller other than main can drive it.
- codeRefs:
  - file:examples/workloads/main.go
  id: r.self-cleaning-temp-dir
  level: MUST
  text: Run creates a temporary directory and removes it on return, and places both
    the materialised runtime binary and the runs directory underneath it.
- codeRefs:
  - file:examples/workloads/main_test.go
  id: r.test-pins-printed-output
  level: MUST
  text: The test drives Run with a captured buffer and fails unless the output contains
    the expected runtime, permissions, argv, terminal states, probe, liveness, allocation
    and in-memory-runner lines.
- codeRefs:
  - file:examples/workloads/main.go
  id: r.ttl-and-joballoc-shown
  level: SHOULD
  text: Run evaluates ttl.Expired for a past completion time and uses joballoc to
    allocate a job's names and report the still-pending subset.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/workloads/main.go` file main.go (examples/workloads/main.go)
- `file:examples/workloads/main_test.go` file main_test.go (examples/workloads/main_test.go)
- `function:4f9e3c5199d0b48e53740bacffb4b9ac` function Run (examples/workloads/main.go)
<!-- SPECD_MANAGED_END -->
