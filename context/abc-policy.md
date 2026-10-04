# Context: abc-policy

Repository: `kcp-libs`

This context exists so that consumers of a policy engine depend on a stable, transport-free contract instead of a concrete HTTP client. It defines the vocabulary and signatures that an implementation such as impl/policyclient must satisfy, and the shape of the task status that callers poll after submission. Keeping the contract in abc/policy lets the engine endpoint stay a per-call argument, so one Client talks to many policy engines.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/policy/policy.go` file policy.go (abc/policy/policy.go)
- `interface:3ec82f13ff355ba5c823fadd91cbc034` interface Client (abc/policy/policy.go)
- `method:60e85e0780bae7128346127fea69a2e3` method Client.Submit (abc/policy/policy.go)
- `method:e1e87944f71f145e385b7bc3db9a5588` method Client.Status (abc/policy/policy.go)
- `struct:c1003284dadd5dbd150449f4f9af4ddc` struct Task (abc/policy/policy.go)
<!-- SPECD_MANAGED_END -->
