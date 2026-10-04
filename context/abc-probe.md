# Context: abc-probe

Repository: `kcp-libs`

This context exists to specify the consecutive-failure probe abstraction: a shared, concurrency-safe place where repeated observations about a named probe key accumulate into a trip/no-trip decision. It is deliberately narrow, holding only counting and threshold normalization, so that callers decide what a key means and what tripping it should do. The behavior encoded here, that a success resets the count, that a new run ID restarts the count, and that keys are independent, is what the tests pin down.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/probe/probe.go` file probe.go (abc/probe/probe.go)
- `file:abc/probe/probe_test.go` file probe_test.go (abc/probe/probe_test.go)
- `function:4fa0354f183a097615e79c3cd44dee84` function EffectiveThreshold (abc/probe/probe.go)
- `function:dd53e30fa995268962209545381294f4` function NewTracker (abc/probe/probe.go)
- `method:3c2f19df0b852dd0e9aafecfd9d631a2` method Tracker.Record (abc/probe/probe.go)
- `method:6bdac2559baf52a89854a4b785722987` method Tracker.Len (abc/probe/probe.go)
- `method:6e9f3e9deecbcb4ec14db06af7386ac7` method Tracker.Forget (abc/probe/probe.go)
- `struct:455c909dc9384fcb23e3388b626fe2ed` struct Tracker (abc/probe/probe.go)
- `struct:c82813d1c5f66d076eef090321523676` struct Counter (abc/probe/probe.go)
<!-- SPECD_MANAGED_END -->
