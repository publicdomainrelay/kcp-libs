# Context: common-logging

Repository: `kcp-libs`

This context exists so that structured logging has one definition instead of every package building its own slog handler. It fixes the decisions that would otherwise drift: output format is JSON, the zero Level means Info rather than Debug, and the service name is a top-level attribute rather than embedded in messages. Discard exists so tests and quiet paths can be handed a real *slog.Logger with no output at all, which keeps signatures uniform and avoids nil-logger branches in callers.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/logging/logging.go` file logging.go (common/logging/logging.go)
- `function:630830e31f658442858bdfe9d070a2bd` function New (common/logging/logging.go)
- `function:7e515b3fbe4a91f9d8b15bdd7d149aca` function Discard (common/logging/logging.go)
- `struct:e5309481f2a5ce717212c88d3d90aa51` struct Options (common/logging/logging.go)
<!-- SPECD_MANAGED_END -->
