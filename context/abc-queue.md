# Context: abc-queue

Repository: `kcp-libs`

This context exists to specify the admission/lease contract of the abc queue: what a Run must look like to be planned, how Policy defaults resolve, how a Capacity, Blocker and reserved count combine into a per-run Admission, and how held leases are scoped to a parent, counted against observed lifecycle phase, expired by TTL and released. It is the durable description of the queue's pure decision functions and its lease bookkeeping, so consumers such as the deno process runner and the admission factory can rely on the same semantics without re-deriving them from the source.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: abc/queue/queue.go
  kind: struct
  name: Admission
  signature: type Admission struct
- file: abc/queue/queue.go
  kind: method
  name: Admission.Waiting
  signature: func (a Admission) Waiting() bool
- file: abc/queue/queue.go
  kind: struct
  name: Blocker
  signature: type Blocker struct
- file: abc/queue/queue.go
  kind: struct
  name: Capacity
  signature: type Capacity struct
- file: abc/queue/queue.go
  kind: function
  name: Decision
  signature: func Decision(policy Policy, maxConcurrent *int32, active, ahead int32)
    (bool, string, string)
- file: abc/queue/queue.go
  kind: function
  name: EffectivePolicy
  signature: func EffectivePolicy(policy Policy) Policy
- file: abc/queue/leases.go
  kind: struct
  name: Lease
  signature: type Lease struct { Parent ref.Ref }
- file: abc/queue/leases.go
  kind: struct
  name: Leases
  signature: type Leases struct { entries *expiringmap.Map[ref.Ref, Lease] }
- file: abc/queue/leases.go
  kind: method
  name: Leases.Count
  signature: func (l *Leases) Count(parent ref.Ref, observed map[ref.Ref]string, now
    time.Time, lifecycle Lifecycle) int32
- file: abc/queue/leases.go
  kind: method
  name: Leases.Forget
  signature: func (l *Leases) Forget(run ref.Ref)
- file: abc/queue/leases.go
  kind: method
  name: Leases.Grant
  signature: func (l *Leases) Grant(run, parent ref.Ref, now time.Time)
- file: abc/queue/leases.go
  kind: method
  name: Leases.Len
  signature: func (l *Leases) Len() int
- file: abc/queue/queue.go
  kind: struct
  name: Lifecycle
  signature: type Lifecycle struct
- file: abc/queue/queue.go
  kind: function
  name: Limit
  signature: func Limit(policy Policy, maxConcurrent *int32) (int32, bool)
- file: abc/queue/leases.go
  kind: function
  name: NewLeases
  signature: func NewLeases(ttl time.Duration) *Leases
- file: abc/queue/queue.go
  kind: function
  name: Order
  signature: func Order(runs []Run)
- file: abc/queue/queue.go
  kind: function
  name: Plan
  signature: func Plan(runs []Run, capacity Capacity, blocker *Blocker, reserved int32,
    lifecycle Lifecycle) []Admission
- file: abc/queue/queue.go
  kind: function
  name: PlanIndex
  signature: func PlanIndex(runs []Run, capacity Capacity, blocker *Blocker, reserved
    int32, lifecycle Lifecycle) map[ref.Ref]Admission
- file: abc/queue/queue.go
  kind: type_alias
  name: Policy
  signature: type Policy = ...
- file: abc/queue/queue.go
  kind: struct
  name: Run
  signature: type Run struct
- file: abc/queue/queue.go
  kind: function
  name: WakeList
  signature: func WakeList(runs []Run, lifecycle Lifecycle, limit int32, unlimited
    bool) []ref.Ref
requirements:
- codeRefs:
  - method:8573365ee09dc1eb4cd105750db7d96a
  - struct:792cc17bc63ccca0dcb38c0a64474daf
  id: r.admission-waiting
  level: MUST
  text: Admission must report through Waiting whether the run it describes is still
    waiting for capacity rather than admitted.
- codeRefs:
  - file:abc/queue/queue.go
  - function:1c23c59b39afb3955d8fb1ef8262bec4
  id: r.decision-verdict
  level: MUST
  text: Decision must take a policy, an optional maxConcurrent override and the active
    and ahead counts, and return whether the run is admitted together with the reason
    strings that explain the verdict.
- codeRefs:
  - file:abc/queue/leases.go
  - method:7fcca607f05203909cae0e30fb95654f
  - method:aaa78ff0e7191ff1df4b246d1a69c083
  - method:ebdc3f4b76becb1059271eec51c5af14
  id: r.leases-grant-forget-len
  level: MUST
  text: Leases must expose Grant to record a run under a parent at a given time, Forget
    to drop the entry for a run, and Len to report the number of live entries.
- codeRefs:
  - file:abc/queue/leases_test.go
  - method:a57993a5f26e81ef1e6ae695a11d798d
  - struct:ad8589e78e7c996950704e3f42b98e35
  id: r.leases-parent-scoped-count
  level: MUST
  text: Leases.Count must count only the leases granted against the given parent ref,
    so a lease held for one parent is never counted for another.
- codeRefs:
  - method:a57993a5f26e81ef1e6ae695a11d798d
  - method:aaa78ff0e7191ff1df4b246d1a69c083
  - struct:34632297f34c7dcc187beda5485fdc49
  id: r.leases-release-on-observed
  level: MUST
  text: A lease must be released when the run it covers is present in the observed
    map and its phase, as read through the Lifecycle, is running or terminal, so that
    Count returns zero and Len drops the entry once both runs are observed.
- codeRefs:
  - file:abc/queue/leases.go
  - function:2ae018e0f488c4c38974fe6ac10ac6b7
  - struct:2f6663c79d6e6a401d99f363c4816759
  id: r.leases-ttl-expiry
  level: MUST
  text: Leases must be created through NewLeases with a TTL, and every entry must
    expire once that TTL has elapsed so that an expired lease no longer contributes
    to the count.
- codeRefs:
  - file:abc/queue/queue.go
  - function:57b793729eefa74795369ef3ec3c5001
  - struct:47c532cd8b7218bfdb153ed7bf19aa12
  id: r.limit-resolution
  level: MUST
  text: Limit must resolve a policy plus an optional maxConcurrent override into an
    effective concurrency limit and a flag saying whether that limit is unlimited.
- codeRefs:
  - function:a64e06cf0280d641c63e9d6a29889ba2
  - function:cb491319026f10dd31ad43cb1a74a3e6
  - struct:4921ee224c96b7635c2b202098cbeff3
  - struct:792cc17bc63ccca0dcb38c0a64474daf
  id: r.plan-admissions
  level: MUST
  text: Plan must take the ordered runs together with capacity, an optional blocker,
    the reserved count and the lifecycle, and return one Admission per run, and PlanIndex
    must expose the same result keyed by ref.Ref.
- codeRefs:
  - file:abc/queue/queue.go
  - function:7ca241acd8161cba27ae96514598782f
  - type_alias:be31cbc6c94318d5c5589f1e679d3f21
  id: r.policy-defaulting
  level: MUST
  text: Policy must be an alias resolved to usable defaults by EffectivePolicy, so
    a caller can pass a partially set or zero policy and receive a complete one.
- codeRefs:
  - file:abc/queue/queue.go
  - file:abc/queue/queue_test.go
  id: r.queue-purity
  level: SHOULD
  text: 'The queue decision functions should stay pure and workload-agnostic: they
    take runs, capacity, policy and lifecycle as arguments and name no workload runtime.'
- codeRefs:
  - file:abc/queue/queue.go
  - function:4a7ac07571d725f388963144fc298480
  - struct:6dd8d61ae12efdab874719004a369656
  id: r.run-ordering
  level: MUST
  text: Order must sort a slice of runs in place into the queue's admission order
    before planning.
- codeRefs:
  - file:abc/queue/queue.go
  - function:7c6d74ae07384500ff8262c7a120089c
  - struct:34632297f34c7dcc187beda5485fdc49
  id: r.wake-list
  level: MUST
  text: WakeList must return the refs of the runs that should be woken given the runs,
    the lifecycle, and either a limit or the unlimited flag.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/queue/leases.go` file leases.go (abc/queue/leases.go)
- `file:abc/queue/leases_test.go` file leases_test.go (abc/queue/leases_test.go)
- `file:abc/queue/queue.go` file queue.go (abc/queue/queue.go)
- `file:abc/queue/queue_test.go` file queue_test.go (abc/queue/queue_test.go)
- `function:1c23c59b39afb3955d8fb1ef8262bec4` function Decision (abc/queue/queue.go)
- `function:2ae018e0f488c4c38974fe6ac10ac6b7` function NewLeases (abc/queue/leases.go)
- `function:4a7ac07571d725f388963144fc298480` function Order (abc/queue/queue.go)
- `function:57b793729eefa74795369ef3ec3c5001` function Limit (abc/queue/queue.go)
- `function:7c6d74ae07384500ff8262c7a120089c` function WakeList (abc/queue/queue.go)
- `function:7ca241acd8161cba27ae96514598782f` function EffectivePolicy (abc/queue/queue.go)
- `function:a64e06cf0280d641c63e9d6a29889ba2` function Plan (abc/queue/queue.go)
- `function:cb491319026f10dd31ad43cb1a74a3e6` function PlanIndex (abc/queue/queue.go)
- `method:7fcca607f05203909cae0e30fb95654f` method Leases.Grant (abc/queue/leases.go)
- `method:8573365ee09dc1eb4cd105750db7d96a` method Admission.Waiting (abc/queue/queue.go)
- `method:a57993a5f26e81ef1e6ae695a11d798d` method Leases.Count (abc/queue/leases.go)
- `method:aaa78ff0e7191ff1df4b246d1a69c083` method Leases.Len (abc/queue/leases.go)
- `method:ebdc3f4b76becb1059271eec51c5af14` method Leases.Forget (abc/queue/leases.go)
- `struct:2f6663c79d6e6a401d99f363c4816759` struct Leases (abc/queue/leases.go)
- `struct:34632297f34c7dcc187beda5485fdc49` struct Lifecycle (abc/queue/queue.go)
- `struct:47c532cd8b7218bfdb153ed7bf19aa12` struct Capacity (abc/queue/queue.go)
- `struct:4921ee224c96b7635c2b202098cbeff3` struct Blocker (abc/queue/queue.go)
- `struct:6dd8d61ae12efdab874719004a369656` struct Run (abc/queue/queue.go)
- `struct:792cc17bc63ccca0dcb38c0a64474daf` struct Admission (abc/queue/queue.go)
- `struct:ad8589e78e7c996950704e3f42b98e35` struct Lease (abc/queue/leases.go)
- `type_alias:be31cbc6c94318d5c5589f1e679d3f21` type_alias Policy (abc/queue/queue.go)
<!-- SPECD_MANAGED_END -->
