# Context: abc-policy

Repository: `kcp-libs`

This context exists so that consumers of a policy engine depend on a stable, transport-free contract instead of a concrete HTTP client. It defines the vocabulary and signatures that an implementation such as impl/policyclient must satisfy, and the shape of the task status that callers poll after submission. Keeping the contract in abc/policy lets the engine endpoint stay a per-call argument, so one Client talks to many policy engines.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/policy/policy.go` file policy.go (abc/policy/policy.go)
<!-- SPECD_MANAGED_END -->
