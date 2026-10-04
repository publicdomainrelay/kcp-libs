# Context: common-logging

Repository: `kcp-libs`

This context exists so that structured logging has one definition instead of every package building its own slog handler. It fixes the decisions that would otherwise drift: output format is JSON, the zero Level means Info rather than Debug, and the service name is a top-level attribute rather than embedded in messages. Discard exists so tests and quiet paths can be handed a real *slog.Logger with no output at all, which keeps signatures uniform and avoids nil-logger branches in callers.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: common/logging/logging.go
  kind: function
  name: Discard
  signature: func Discard() *slog.Logger
- file: common/logging/logging.go
  kind: function
  name: New
  signature: func New(opts Options) *slog.Logger
- file: common/logging/logging.go
  kind: struct
  name: Options
  signature: type Options struct { Service string; Writer io.Writer; Level slog.Level
    }
requirements:
- codeRefs:
  - file:common/logging/logging.go
  - function:7e515b3fbe4a91f9d8b15bdd7d149aca
  id: r.discard-writes-nothing
  level: MUST
  text: Discard must return a *slog.Logger whose handler writes to io.Discard at slog.LevelError,
    so callers get a usable logger that emits no output.
- codeRefs:
  - file:common/logging/logging.go
  - function:630830e31f658442858bdfe9d070a2bd
  - struct:e5309481f2a5ce717212c88d3d90aa51
  id: r.new-defaults-level-to-info
  level: MUST
  text: New must replace a zero opts.Level with slog.LevelInfo, so an unset level
    means Info and not the zero value of slog.Level.
- codeRefs:
  - function:630830e31f658442858bdfe9d070a2bd
  - struct:e5309481f2a5ce717212c88d3d90aa51
  id: r.new-tags-records-with-the-service
  level: MUST
  text: New must attach a "service" attribute to the logger only when opts.Service
    is not empty; an empty Service must leave the logger untagged.
- codeRefs:
  - function:630830e31f658442858bdfe9d070a2bd
  - struct:e5309481f2a5ce717212c88d3d90aa51
  id: r.new-writes-json-to-the-writer
  level: MUST
  text: New must build the logger on a slog.NewJSONHandler over opts.Writer with the
    resolved level, so records are JSON lines on the writer the caller supplied.
- codeRefs:
  - file:common/logging/logging.go
  - function:630830e31f658442858bdfe9d070a2bd
  - struct:e5309481f2a5ce717212c88d3d90aa51
  id: r.options-is-the-only-wiring
  level: SHOULD
  text: Loggers should be produced only through Options via New or Discard, so format,
    default level, and the service attribute stay in one place.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/logging/logging.go` file logging.go (common/logging/logging.go)
- `function:630830e31f658442858bdfe9d070a2bd` function New (common/logging/logging.go)
- `function:7e515b3fbe4a91f9d8b15bdd7d149aca` function Discard (common/logging/logging.go)
- `struct:e5309481f2a5ce717212c88d3d90aa51` struct Options (common/logging/logging.go)
<!-- SPECD_MANAGED_END -->
