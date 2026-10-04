# Context: abc-joballoc

Repository: `kcp-libs`

This context exists so the abc layer has one place that answers which names a job already holds and which of those a controller has not yet seen, without each consumer keeping its own bookkeeping. It is deliberately runtime-agnostic: it names no workload runtime and depends only on common/ref and common/expiringmap, so any reconciler in the repo can use it. Expiry comes from the shared expiringmap rather than from a timer here, which keeps the allocator free of goroutines and makes every read testable against an explicit now.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: abc/joballoc/joballoc.go
  kind: struct
  name: Allocation
  signature: type Allocation struct { Name string; At time.Time }
- file: abc/joballoc/joballoc.go
  kind: struct
  name: Allocator
  signature: type Allocator struct { byJob *expiringmap.Map[ref.Ref, []Allocation]
    }
- file: abc/joballoc/joballoc.go
  kind: method
  name: Allocator.Allocate
  signature: func (a *Allocator) Allocate(r ref.Ref, names []string, now time.Time)
- file: abc/joballoc/joballoc.go
  kind: method
  name: Allocator.Forget
  signature: func (a *Allocator) Forget(r ref.Ref)
- file: abc/joballoc/joballoc.go
  kind: method
  name: Allocator.Len
  signature: func (a *Allocator) Len() int
- file: abc/joballoc/joballoc.go
  kind: method
  name: Allocator.Names
  signature: func (a *Allocator) Names(r ref.Ref, now time.Time) []string
- file: abc/joballoc/joballoc.go
  kind: method
  name: Allocator.Pending
  signature: func (a *Allocator) Pending(r ref.Ref, observed []string, now time.Time)
    []string
- file: abc/joballoc/joballoc.go
  kind: function
  name: MergeNames
  signature: func MergeNames(groups ...[]string) []string
- file: abc/joballoc/joballoc.go
  kind: function
  name: New
  signature: func New(ttl time.Duration) *Allocator
requirements:
- codeRefs:
  - method:1d809a0f9cc0db66dd09568aab52e10e
  - struct:c6c45587967912c642f434fc5c395fac
  id: r.allocate-appends-and-skips-empty
  level: MUST
  text: 'Allocator.Allocate returns immediately when names is empty, and otherwise
    reads the existing allocations for the ref and appends one Allocation{Name, At:
    now} per name before setting the result back.'
- codeRefs:
  - method:10aee9a9b59c82c429e59eda7063cff2
  id: r.forget-deletes-the-job
  level: MUST
  text: Allocator.Forget deletes the ref's entry from the underlying map.
- codeRefs:
  - method:99144fd4f43e31bdffe2206a3a99a050
  id: r.len-reports-held-jobs
  level: MUST
  text: Allocator.Len returns the number of entries the underlying map currently holds.
- codeRefs:
  - function:3f2d8c8fe1fe9a1ffe2aec6ae9cb21b1
  id: r.merge-names-order-preserving-dedupe
  level: MUST
  text: MergeNames concatenates the given groups in order, dropping empty names and
    names already emitted, so the result keeps first-seen order with no duplicates.
- codeRefs:
  - method:2fe2fe5409de2aff6260db2ad57424e1
  id: r.names-nil-when-absent-or-empty
  level: MUST
  text: Allocator.Names returns nil when the ref has no live entry or the entry holds
    no allocations, and otherwise returns the allocation names in stored order.
- codeRefs:
  - file:abc/joballoc/joballoc.go
  - function:e5835aa23953e8fe5e8ba0521c6f6305
  - struct:81a3d6c6679a07f9c85fcbb079cd24ef
  id: r.new-wires-expiring-map-with-ttl
  level: MUST
  text: New returns an Allocator whose byJob map is an expiringmap.Map keyed by ref.Ref
    holding []Allocation, created with the given ttl.
- codeRefs:
  - method:9c2e809f72d7951e74cf7925860610ec
  id: r.pending-excludes-observed-and-dedupes
  level: MUST
  text: Allocator.Pending returns nil when the ref has no live entry, and otherwise
    returns the allocated names not present in observed, skipping any name already
    emitted so the result carries no duplicates.
- codeRefs:
  - method:1d809a0f9cc0db66dd09568aab52e10e
  - method:2fe2fe5409de2aff6260db2ad57424e1
  - method:9c2e809f72d7951e74cf7925860610ec
  id: r.reads-honour-explicit-now
  level: MUST
  text: Every read and write takes an explicit now time.Time and passes it to the
    expiringmap, so expiry is decided by the caller's clock and not by the allocator.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/joballoc/joballoc.go` file joballoc.go (abc/joballoc/joballoc.go)
- `file:abc/joballoc/joballoc_test.go` file joballoc_test.go (abc/joballoc/joballoc_test.go)
- `function:3f2d8c8fe1fe9a1ffe2aec6ae9cb21b1` function MergeNames (abc/joballoc/joballoc.go)
- `function:e5835aa23953e8fe5e8ba0521c6f6305` function New (abc/joballoc/joballoc.go)
- `method:10aee9a9b59c82c429e59eda7063cff2` method Allocator.Forget (abc/joballoc/joballoc.go)
- `method:1d809a0f9cc0db66dd09568aab52e10e` method Allocator.Allocate (abc/joballoc/joballoc.go)
- `method:2fe2fe5409de2aff6260db2ad57424e1` method Allocator.Names (abc/joballoc/joballoc.go)
- `method:99144fd4f43e31bdffe2206a3a99a050` method Allocator.Len (abc/joballoc/joballoc.go)
- `method:9c2e809f72d7951e74cf7925860610ec` method Allocator.Pending (abc/joballoc/joballoc.go)
- `struct:81a3d6c6679a07f9c85fcbb079cd24ef` struct Allocator (abc/joballoc/joballoc.go)
- `struct:c6c45587967912c642f434fc5c395fac` struct Allocation (abc/joballoc/joballoc.go)
<!-- SPECD_MANAGED_END -->
