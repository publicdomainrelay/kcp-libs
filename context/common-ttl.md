# Context: common-ttl

Repository: `kcp-libs`

This context exists so callers do not each re-derive TTL semantics from raw pointers. Kubernetes-style APIs express TTLs as optional seconds on a timestamp, where nil means unset and negative is meaningless; these helpers centralize that normalization so every consumer agrees on when something is scheduled for deletion, and when nothing should be scheduled at all.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: common/ttl/ttl.go
  kind: function
  name: Deadline
  signature: func Deadline(start *metav1.Time, seconds *int64, now time.Time) bool
- file: common/ttl/ttl.go
  kind: function
  name: Effective
  signature: func Effective(override, fallback *int64) *int64
- file: common/ttl/ttl.go
  kind: function
  name: Expired
  signature: func Expired(completion *metav1.Time, seconds *int64, now time.Time)
    (expired bool, after time.Duration, known bool)
- file: common/ttl/ttl.go
  kind: function
  name: Expiry
  signature: func Expiry(completion *metav1.Time, seconds *int64) (time.Time, bool)
requirements:
- codeRefs:
  - file:common/ttl/ttl.go
  - function:9fdc74d33c3a88840f75732e3c0170bb
  id: r.deadline-unknown-without-start
  level: MUST
  text: Deadline returns false when start is nil, seconds is nil, or seconds is negative.
    Otherwise it returns true only when now is not before start plus seconds seconds.
- codeRefs:
  - file:common/ttl/ttl.go
  - function:d687afb3744f79be091b78a990794a9d
  id: r.effective-prefers-override
  level: MUST
  text: Effective returns the value of override when override is not nil, otherwise
    the value of fallback. When the selected value is nil or negative, Effective returns
    nil. When it is non-negative, Effective returns a pointer to a copy of the value,
    never an alias of the input.
- codeRefs:
  - function:503ea477e92bcc1bc7f9bdba99922323
  - function:a0c1200f8b55a18bfa918f1f0cee7615
  id: r.expired-reports-unknown
  level: MUST
  text: Expired returns (false, 0, false) when no expiry is known, that is when Expiry
    reports false. When the expiry is known and now is not before it, Expired returns
    (true, 0, true). When the expiry is in the future, Expired returns (false, expiry.Sub(now),
    true).
- codeRefs:
  - file:common/ttl/ttl.go
  - function:503ea477e92bcc1bc7f9bdba99922323
  id: r.expiry-needs-completion-and-seconds
  level: MUST
  text: Expiry returns the zero time and false when completion is nil, seconds is
    nil, or seconds is negative. Otherwise it returns completion plus seconds seconds
    and true.
- codeRefs:
  - function:503ea477e92bcc1bc7f9bdba99922323
  - function:9fdc74d33c3a88840f75732e3c0170bb
  - function:d687afb3744f79be091b78a990794a9d
  id: r.negative-seconds-mean-unset
  level: MUST
  text: 'A negative seconds value is treated exactly like an unset one across Effective,
    Expiry and Deadline: no expiry and no deadline are produced.'
- codeRefs:
  - file:common/ttl/ttl_test.go
  - function:9fdc74d33c3a88840f75732e3c0170bb
  - function:a0c1200f8b55a18bfa918f1f0cee7615
  id: r.tests-cover-boundaries
  level: SHOULD
  text: The tests cover override preference and negative rejection in Effective, a
    future expiry, a past expiry and a missing completion time in Expired, and a future
    deadline, a reached deadline and a missing start time in Deadline.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/ttl/ttl.go` file ttl.go (common/ttl/ttl.go)
- `file:common/ttl/ttl_test.go` file ttl_test.go (common/ttl/ttl_test.go)
- `function:503ea477e92bcc1bc7f9bdba99922323` function Expiry (common/ttl/ttl.go)
- `function:9fdc74d33c3a88840f75732e3c0170bb` function Deadline (common/ttl/ttl.go)
- `function:a0c1200f8b55a18bfa918f1f0cee7615` function Expired (common/ttl/ttl.go)
- `function:d687afb3744f79be091b78a990794a9d` function Effective (common/ttl/ttl.go)
<!-- SPECD_MANAGED_END -->
