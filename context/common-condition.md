# Context: common-condition

Repository: `kcp-libs`

This context exists to specify the shared condition helpers that kcp-libs controllers use when reporting status. Rather than each controller reimplementing condition bookkeeping, common/condition centralises it: a single Set entry point that records type, status, reason, message and observed generation through meta.SetStatusCondition, thin SetTrue/SetFalse wrappers over that entry point, and the small reader/eraser functions Of, Is, Remove and Copy that callers need around it. It exists so condition handling is uniform, so the observed generation is never forgotten, and so callers can hand out copies of a condition list without aliasing the live one.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
