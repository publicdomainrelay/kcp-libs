# Context: common-clientlimit

Repository: `kcp-libs`

This context exists because client-go defaults every client to 5 qps with a burst of 10, which starves a controller that watches and reconciles many objects. The package centralises the fix in one function so each controller construction path (exportwatch, informerwatch, kcpstore) applies the same rate-limit policy without duplicating the defaulting logic, and keeps caller intent: an operator who explicitly sets QPS or Burst on a config is not overridden. It sits in the generic common layer, names no workload runtime, and is usable by any component that builds a rest.Config.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: common/clientlimit/clientlimit.go
  kind: function
  name: Apply
  signature: func Apply(cfg *rest.Config, qps float32, burst int) *rest.Config
requirements:
- codeRefs:
  - file:common/clientlimit/clientlimit.go
  - function:67bc041de773d39e4803a9795f096d96
  id: r.caller-value-wins
  level: MUST
  text: Apply must set out.QPS only when the copied config's QPS is <= 0 and out.Burst
    only when its Burst is <= 0, so a rate limit the caller already set on the config
    is kept.
- codeRefs:
  - file:common/clientlimit/clientlimit_test.go
  id: r.default-above-client-go-default
  level: MUST
  text: DefaultQPS must stay greater than 5, because client-go defaults every client
    to 5 qps with a burst of 10, which starves a controller.
- codeRefs:
  - file:common/clientlimit/clientlimit_test.go
  - function:67bc041de773d39e4803a9795f096d96
  id: r.explicit-option-overrides-empty-config
  level: MUST
  text: When the config carries no rate limit, a positive qps and burst argument must
    be used as the resulting QPS and Burst.
- codeRefs:
  - file:common/clientlimit/clientlimit.go
  id: r.generic-layer-no-runtime
  level: SHOULD
  text: 'The package must stay generic: it names no workload runtime and works from
    a plain *rest.Config, so any controller construction path can use it.'
- codeRefs:
  - file:common/clientlimit/clientlimit.go
  - function:67bc041de773d39e4803a9795f096d96
  id: r.non-positive-option-falls-back-to-default
  level: MUST
  text: A qps less than or equal to zero must fall back to DefaultQPS, and a burst
    less than or equal to zero must fall back to DefaultBurst.
- codeRefs:
  - file:common/clientlimit/clientlimit.go
  - function:67bc041de773d39e4803a9795f096d96
  id: r.returns-copy-not-mutation
  level: MUST
  text: Apply must copy the given config with rest.CopyConfig and return the copy;
    the config passed by the caller must not be mutated.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/clientlimit/clientlimit.go` file clientlimit.go (common/clientlimit/clientlimit.go)
- `file:common/clientlimit/clientlimit_test.go` file clientlimit_test.go (common/clientlimit/clientlimit_test.go)
- `function:67bc041de773d39e4803a9795f096d96` function Apply (common/clientlimit/clientlimit.go)
<!-- SPECD_MANAGED_END -->
