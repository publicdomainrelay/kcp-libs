# Context: abc-queue

Repository: `kcp-libs`

This context exists to specify the admission/lease contract of the abc queue: what a Run must look like to be planned, how Policy defaults resolve, how a Capacity, Blocker and reserved count combine into a per-run Admission, and how held leases are scoped to a parent, counted against observed lifecycle phase, expired by TTL and released. It is the durable description of the queue's pure decision functions and its lease bookkeeping, so consumers such as the deno process runner and the admission factory can rely on the same semantics without re-deriving them from the source.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
