# Context: common-denocomputer

Repository: `kcp-libs`

This context exists to keep one copy of the deno.computer names and phases that both this generic library and the deno-kcp consumer agree on. Rather than letting the library invent its own group, finalizers, labels, conditions or phase strings, the constants live here and a test reads the consumer's own source as the source of truth and fails on any drift. The phase predicates sit beside the Phase type so callers in the reconcile paths can ask whether a run, pod, job, policy workflow or policy engine has finished without re-encoding which phases count as terminal in every call site.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/denocomputer/denocomputer.go` file denocomputer.go (common/denocomputer/denocomputer.go)
- `file:common/denocomputer/denocomputer_test.go` file denocomputer_test.go (common/denocomputer/denocomputer_test.go)
- `function:0ac0b313ea9228a561bfd9f56465b4be` function TerminalPolicyWorkflow (common/denocomputer/denocomputer.go)
- `function:0d13333d7202e69c40af4472e0f2436b` function TerminalDenoPod (common/denocomputer/denocomputer.go)
- `function:36ff8f97f3450e32e33a6b3b0e493610` function TerminalDenoJob (common/denocomputer/denocomputer.go)
- `function:67ee37b108a01a226ce8f48241d735b8` function RunningPhase (common/denocomputer/denocomputer.go)
- `function:6e7c1837ac42c5ecd27e997607373d2d` function TerminalDenoRun (common/denocomputer/denocomputer.go)
- `function:8947f352a45e5a31be8d9bdad53c8a93` function TerminalPolicyEngine (common/denocomputer/denocomputer.go)
- `type_alias:a265ac7a2b0813430059371d42a76985` type_alias Phase (common/denocomputer/denocomputer.go)
<!-- SPECD_MANAGED_END -->
