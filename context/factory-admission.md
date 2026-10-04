# Context: factory-admission

Repository: `kcp-libs`

This context exists to describe the admission gate that sits between a queue and the workloads it schedules. It defines the Source port through which the gate reads cluster state (parent, siblings, capacity) without knowing any workload runtime, and the Admitter that turns those reads plus a lease table into a queue.Admission decision and into wake notifications for runs that have become eligible. It is the piece that enforces per-parent concurrency policy, so callers can ask whether a specific run may proceed and can be told which runs to wake once capacity frees up.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:factory/admission/admission.go` file admission.go (factory/admission/admission.go)
- `file:factory/admission/admission_test.go` file admission_test.go (factory/admission/admission_test.go)
- `function:4204d25ce951cefafd1bb89729e7bd03` function New (factory/admission/admission.go)
- `interface:9c23d5af968c6dcd87d996b5fb785c81` interface Source (factory/admission/admission.go)
- `method:0d37cff2d12a3e38aa4393696696d4c5` method Source.Runs (factory/admission/admission.go)
- `method:596d1d41bb4565e406b326020ce25a82` method Source.Parent (factory/admission/admission.go)
- `method:9b26f8dc3db4f420718a3ae685d15ea8` method Admitter.Admit (factory/admission/admission.go)
- `method:d3904e1762452936d94101cc4cfc4bbe` method Source.Capacity (factory/admission/admission.go)
- `method:e452266abf564c3d1e61f18fb3d426f6` method Admitter.Leases (factory/admission/admission.go)
- `method:fb3c595a96a98de51678ad45702f4519` method Admitter.Wake (factory/admission/admission.go)
- `struct:3aede7d4f92074172612567f434a4408` struct Options (factory/admission/admission.go)
- `struct:c62339005d8ea8281d87db8bcb56a638` struct Admitter (factory/admission/admission.go)
<!-- SPECD_MANAGED_END -->
