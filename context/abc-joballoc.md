# Context: abc-joballoc

Repository: `kcp-libs`

This context exists so the abc layer has one place that answers which names a job already holds and which of those a controller has not yet seen, without each consumer keeping its own bookkeeping. It is deliberately runtime-agnostic: it names no workload runtime and depends only on common/ref and common/expiringmap, so any reconciler in the repo can use it. Expiry comes from the shared expiringmap rather than from a timer here, which keeps the allocator free of goroutines and makes every read testable against an explicit now.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
