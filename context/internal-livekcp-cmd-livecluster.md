# Context: internal-livekcp-cmd-livecluster

Repository: `kcp-libs`

The command exists so a developer or test harness can start a real kcp cluster as a child process and learn its coordinates from a file, instead of embedding cluster startup in Go test code. It is deliberately thin: all cluster logic lives in the livekcp package, and this file only handles flags, process signals, environment-file emission, and exit codes.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
requirements:
- codeRefs:
  - file:internal/livekcp/cmd/livecluster/main.go
  id: r.atomic-env-file
  level: MUST
  text: The command writes `export KCP_LIBS_KUBECONFIG=<kubeconfig>` and `export KCP_LIBS_SERVER=<server>`
    to the env file by writing a sibling `--env` path suffixed with `.tmp` and renaming
    it into place, exiting with status 1 on either failure.
- codeRefs:
  - file:internal/livekcp/cmd/livecluster/main.go
  id: r.env-flag-required
  level: MUST
  text: 'The command accepts a `--env` flag naming the file to write the environment
    to, and exits with status 2 and the message `livecluster: --env is required` on
    stderr when it is empty.'
- codeRefs:
  - file:internal/livekcp/cmd/livecluster/main.go
  id: r.print-trace-then-block
  level: MUST
  text: After the env file is in place, the command prints the cluster trace to stdout
    and blocks until the context is done.
- codeRefs:
  - file:internal/livekcp/cmd/livecluster/main.go
  id: r.start-cluster-from-signal-context
  level: MUST
  text: The command starts the cluster with a context that is cancelled by SIGINT
    or SIGTERM, and exits with status 1 and the error prefix `livecluster:` on stderr
    when startup fails.
- codeRefs:
  - file:internal/livekcp/cmd/livecluster/main.go
  id: r.stop-cluster-on-exit
  level: MUST
  text: The command defers Stop on the started cluster so the cluster is torn down
    when the process leaves main.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:internal/livekcp/cmd/livecluster/main.go` file main.go (internal/livekcp/cmd/livecluster/main.go)
<!-- SPECD_MANAGED_END -->
