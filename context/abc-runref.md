# Context: abc-runref

Repository: `kcp-libs`

This context exists so a reconciler can tell a fresh start from a duplicate of a run that is already in flight. Records live only for the configured TTL, so the index is a cache of recent starts rather than durable state, and callers are expected to refresh a record on each pass (Keep) and drop it when the run reaches a terminal phase. AlreadyStarted encodes the refusal rule the admission path needs, including the cases that must NOT be refused: no record, the same runID, a run that already started, a retry, or a recreated object that carries a different UID behind the same name.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/runref/runref.go` file runref.go (abc/runref/runref.go)
- `file:abc/runref/runref_test.go` file runref_test.go (abc/runref/runref_test.go)
- `function:02b977ff8ed4f89bd7fb19f709f855a0` function New (abc/runref/runref.go)
- `function:b3be3bdb00e0f276639e92a391046532` function AlreadyStarted (abc/runref/runref.go)
- `method:31e9e3f14991919b8dc8c86a1ae20707` method Index.Record (abc/runref/runref.go)
- `method:a092ed399600f5b0fda4c653bab06826` method Index.Lookup (abc/runref/runref.go)
- `method:cae442026fb66407a9804c8987ffee8e` method Index.Forget (abc/runref/runref.go)
- `method:f29af6e54699815a89e6dfded48853b8` method Index.Keep (abc/runref/runref.go)
- `method:f61f5a9ef40992f65bfc79298b0fc1b8` method Index.Len (abc/runref/runref.go)
- `struct:27d20ba09467d3ee31db323ffe88bc2d` struct Index (abc/runref/runref.go)
- `struct:47cf33792836866fa369462cf758f59b` struct Current (abc/runref/runref.go)
- `struct:f9c19c773fc6457d35bf1a48e8d0750c` struct Record (abc/runref/runref.go)
<!-- SPECD_MANAGED_END -->
