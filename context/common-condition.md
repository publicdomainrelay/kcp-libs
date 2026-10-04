# Context: common-condition

Repository: `kcp-libs`

This context exists to specify the shared condition helpers that kcp-libs controllers use when reporting status. Rather than each controller reimplementing condition bookkeeping, common/condition centralises it: a single Set entry point that records type, status, reason, message and observed generation through meta.SetStatusCondition, thin SetTrue/SetFalse wrappers over that entry point, and the small reader/eraser functions Of, Is, Remove and Copy that callers need around it. It exists so condition handling is uniform, so the observed generation is never forgotten, and so callers can hand out copies of a condition list without aliasing the live one.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/condition/condition.go` file condition.go (common/condition/condition.go)
- `file:common/condition/condition_test.go` file condition_test.go (common/condition/condition_test.go)
- `function:2f22991378805fd945c1a0d94b1e9b0e` function Of (common/condition/condition.go)
- `function:741af730ad7d212947f2d9d42bf55557` function Copy (common/condition/condition.go)
- `function:a5845019d1d1a5d547885188f6cf2d6a` function SetTrue (common/condition/condition.go)
- `function:d9d1ab909c15980c4699a3cb56bea924` function Set (common/condition/condition.go)
- `function:e1d40351c0eded29676ce645d7c908ab` function Is (common/condition/condition.go)
- `function:ead30502c3fe49e900a144e9e71219a5` function Remove (common/condition/condition.go)
- `function:eee1a190d5df1941cf6cb5d346a99684` function SetFalse (common/condition/condition.go)
<!-- SPECD_MANAGED_END -->
