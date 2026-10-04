# Context: examples-policy

Repository: `kcp-libs`

This context documents the examples/policy package: the runnable demonstration of the policy engine surface. It exists so the exported behaviour of that example — the engine's URL accessor, its Close, its submitted-run counter, its HTTP handler, and the Run entry point — is described and anchored to code, letting a reader or a tool reason about what the example guarantees without reading every line.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/policy/engine.go` file engine.go (examples/policy/engine.go)
- `file:examples/policy/main.go` file main.go (examples/policy/main.go)
- `file:examples/policy/main_test.go` file main_test.go (examples/policy/main_test.go)
- `function:4cf3070e763136f1251e403942485497` function Run (examples/policy/main.go)
- `method:5e7dd0b9cb49339a8dce6360d81790ab` method engine.Submits (examples/policy/engine.go)
- `method:8adc047283ab850bffddf2fa6d30a23c` method engine.URL (examples/policy/engine.go)
- `method:8efb924d3babc92148f8bda3b2d45469` method engine.Close (examples/policy/engine.go)
- `method:dd19ad85f3bd4cb7b58e6d21d7d16180` method engine.ServeHTTP (examples/policy/engine.go)
<!-- SPECD_MANAGED_END -->
