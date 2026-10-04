# Context: common-logging

Repository: `kcp-libs`

This context exists so that structured logging has one definition instead of every package building its own slog handler. It fixes the decisions that would otherwise drift: output format is JSON, the zero Level means Info rather than Debug, and the service name is a top-level attribute rather than embedded in messages. Discard exists so tests and quiet paths can be handed a real *slog.Logger with no output at all, which keeps signatures uniform and avoids nil-logger branches in callers.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
