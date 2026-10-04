# Context: examples-admission

Repository: `kcp-libs`

This context exists so that the admission example can be described and regenerated as a compile-checked demonstration of the admission Source interface. It documents the exact contract an implementer must satisfy: how a parent is discovered from a run (a label lookup that may legitimately find nothing), how child runs are enumerated and filtered by that same label, and how capacity is derived from the parent batch, including the not-found case which must surface as a Blocker rather than an error. Keeping the spec anchored to these methods lets changes to the queue capacity and blocker types be validated against a real consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/admission/main.go` file main.go (examples/admission/main.go)
- `file:examples/admission/main_test.go` file main_test.go (examples/admission/main_test.go)
- `function:382c144be35e7a3b04857018760f4c2b` function Run (examples/admission/main.go)
- `method:4b1c829db1dd2db486e75edae8ec99f8` method source.Capacity (examples/admission/main.go)
- `method:9391ce08b0aaa1ca190f09656b2de077` method source.Parent (examples/admission/main.go)
- `method:e222d2ede890130c86294ba108430575` method source.Runs (examples/admission/main.go)
<!-- SPECD_MANAGED_END -->
