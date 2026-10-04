# Context: common-statuspatch

Repository: `kcp-libs`

This context exists to specify the reusable patch-construction vocabulary shared by every reconcile and store layer in kcp-libs. It keeps the generic JSON-patch byte shapes in one place so controllers do not hand-roll status envelopes, resourceVersion stamps, or finalizer patches, and so the exact wire shape of those patches is pinned by unit tests rather than duplicated per consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/statuspatch/statuspatch.go` file statuspatch.go (common/statuspatch/statuspatch.go)
- `file:common/statuspatch/statuspatch_test.go` file statuspatch_test.go (common/statuspatch/statuspatch_test.go)
- `function:146b92aee570e1cccf066680c85a4988` function Merge (common/statuspatch/statuspatch.go)
- `function:973dee6f0664f3fb6b8b0860b5ec3a25` function FinalizerAdd (common/statuspatch/statuspatch.go)
- `function:a0ef10c6d57cf09c7ea23871b9b2c059` function FinalizersOf (common/statuspatch/statuspatch.go)
- `function:e490ecee35ab297d47aa943e045c98b3` function WithResourceVersion (common/statuspatch/statuspatch.go)
- `function:e6682cb9b4a0a03a95c84fedefb98805` function FinalizerRemove (common/statuspatch/statuspatch.go)
- `struct:9ddae76e6ab972a834ddad24d5320169` struct Metadata (common/statuspatch/statuspatch.go)
<!-- SPECD_MANAGED_END -->
