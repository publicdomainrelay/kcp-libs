# Context: common-statuspatch

Repository: `kcp-libs`

This context exists to specify the reusable patch-construction vocabulary shared by every reconcile and store layer in kcp-libs. It keeps the generic JSON-patch byte shapes in one place so controllers do not hand-roll status envelopes, resourceVersion stamps, or finalizer patches, and so the exact wire shape of those patches is pinned by unit tests rather than duplicated per consumer.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
