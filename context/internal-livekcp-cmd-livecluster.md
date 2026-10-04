# Context: internal-livekcp-cmd-livecluster

Repository: `kcp-libs`

The command exists so a developer or test harness can start a real kcp cluster as a child process and learn its coordinates from a file, instead of embedding cluster startup in Go test code. It is deliberately thin: all cluster logic lives in the livekcp package, and this file only handles flags, process signals, environment-file emission, and exit codes.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:internal/livekcp/cmd/livecluster/main.go` file main.go (internal/livekcp/cmd/livecluster/main.go)
<!-- SPECD_MANAGED_END -->
