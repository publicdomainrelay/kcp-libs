# kcp-libs

Reusable Go libraries for kcp controllers, extracted from `deno-kcp`. The
module is `github.com/publicdomainrelay/kcp-libs`.

The org-root ABC layering pattern, translated to idiomatic Go: one module, one
package per concept-layer, and a dependency arrow that only points one way.
No comments in the code; a package's path and names carry the meaning.

```
common/    leaf: types, constants, pure helpers            external deps only
   ^
abc/       interfaces and pure state, no I/O               common only
   ^
impl/      concrete bindings: client-go, os/exec, net/http abc + common
   ^
factory/   composition: driver, admission, wiring          impl + abc + common
   ^
examples/  runnable programs, one per use case             anything

internal/  test support: a live kcp, and the import rules   not for production
```

**Start at [`examples/`](examples/README.md).** Six runnable programs drive the
library, with a table mapping every package to the situation it is for. Three
of them talk to a real kcp and need `kcp` and `kubectl` on PATH, which is also
what the live tests need; the other three need nothing but Go. `go run
./examples/controller` is the whole reconcile loop in one file.

`internal/boundaries` is a test, not a package: it reads `go list -json ./...`
and fails the build if any package imports against the arrow, if a `common`
package imports anything project-local, or if an `abc` package imports past
`common`.

## Layout

| Layer | Package | What it is |
|---|---|---|
| common | `common/ref` | `Ref{LogicalCluster, Namespace, Name, ResourceVersion}`, composite keys, `BaseHost`/`ClusterURL` |
| common | `common/kcp` | `kcp.io/cluster`, `kcp.io/path`, service labels and FQDNs, the inverse |
| common | `common/condition` | `metav1.Condition` set/remove/find, transition-time preservation |
| common | `common/statuspatch` | merge-patch bodies, resource-version stamping, finalizer JSON patches, `Optional` |
| common | `common/deno` | the `deno.computer` domain: labels, finalizers, conditions, phases, terminal predicates, concurrency policies |
| common | `common/denospec` | the shared wire shape: pod template, exec probe, service account ref, permissions, deno argv |
| common | `common/ttl` | retention and active-deadline decisions |
| common | `common/outputs` | `map[string]any` to `map[string]string` |
| common | `common/logging`, `common/env` | JSON slog logger; env-or-default readers |
| abc | `abc/reconcile` | the decider seam: `Result[Status]`, operations, `Reconciler[Observed, Status]` |
| abc | `abc/queue` | **the queue semantics**: capacity, `Decision`, `Plan`, `PlanIndex`, `PlanByParent`, `WakeList`, `Leases` |
| abc | `abc/driver` | work `Key`, `Handler`, `Queue`, and the requeue `Policy` |
| abc | `abc/cache` | `Indexer`/`Set`, index names and index functions, `Decode[T]` |
| abc | `abc/probe` | liveness failure counters and the threshold decision |
| abc | `abc/runref` | the in-memory runID-to-ref index, and the duplicate-start guard |
| abc | `abc/joballoc` | created-but-unobserved run names, held until the status write lands |
| abc | `abc/runner` | `PodRunner`/`EngineRunner` and their request/status types |
| abc | `abc/pki` | the OpenBao PKI client port, certificate types, `Provisioner` |
| abc | `abc/store` | generic `Reader[T]`/`Writer[T]`/`Resource[T]`, token minter, path resolver |
| abc | `abc/policy` | the policy engine client port |
| impl | `impl/kcpstore` | client-go REST store: generic typed resources, status patches, finalizers, token minting, cluster paths |
| impl | `impl/exportwatch` | APIExportEndpointSlice discovery and the await loop for virtual workspace URLs |
| impl | `impl/informerwatch` | shared dynamic informers over `/clusters/*`, indexers, event handlers |
| impl | `impl/execrunner` | `os/exec` pod and engine runners: run directories, process groups, recovery, probes |
| impl | `impl/memoryrunner` | in-memory runners for tests |
| impl | `impl/openbaoclient` | a typed adapter over the official `github.com/openbao/openbao/api/v2` client |
| impl | `impl/pkiprovisioner` | one intermediate CA per namespace, root in the root namespace, cached |
| impl | `impl/policyclient` | the gha-lite policy engine HTTP client, including verdict extraction |
| impl | `impl/metrics` | dependency-free Prometheus text registry |
| impl | `impl/assets` | writes a caller's asset set next to a runs directory, once |
| factory | `controller` | informers + workqueue + worker pool + requeue policy + metrics |
| factory | `admission` | per-parent admission: leases, planning, and the wake of queued runs |
| factory | `dns` | the FQDN table, a pod's own name, and one token per workspace |
| examples | `examples/*` | one runnable program per use case, each asserted by its own test |
| support | `internal/livekcp` | starts a real kcp, applies a schema and an export, binds consumer workspaces |

## The two pieces worth reading first

**`abc/queue`** is the per-parent admission queue. A parent (a
`PolicyWorkflowPod`) caps how many children (runs) execute at once.
`Capacity`/`Decision` decide one run; `Plan` decides a whole parent from an
oldest-first ordering and blocks nothing when the parent is missing or not
ready; `PlanIndex` answers by run ref, which is what a reconciler has.
`Leases` is the in-memory guard for a run that has been admitted but not yet
observed `Running`: without it several pending runs see `active=0` in the same
window and the cap leaks. `WakeList` is what a freed slot enqueues.

**`abc/driver`** is the requeue policy. `Policy.Next` reproduces the measured
behaviour: a terminal key with no requeue stops, an unset requeue becomes the
interval, and a key whose kind is clamped is re-checked at
`MinTransitionPoll` because a running process is only visible by asking it.
`factory/controller` runs the loop: one worker per key, conflicts re-enqueue
rather than overwrite, errors go through the rate limiter.

## Refactor map

Where the code in `../deno-kcp` moves.

| deno-kcp | kcp-libs |
|---|---|
| `internal/provider/registry.go`, `registry_*.go`, `cluster_path.go` | `impl/kcpstore`, `common/ref`, `common/statuspatch` |
| `internal/provider/watch.go`, `watch_cache.go`, `driver.go` | `impl/informerwatch`, `impl/exportwatch`, `abc/cache`, `abc/driver`, `factory/controller` |
| `internal/provider/admission.go` | `abc/queue`, `factory/admission` |
| `internal/provider/metrics.go` | `impl/metrics` |
| `internal/provider/service_dns.go` | `common/kcp`, `factory/dns` (name and token table composition) |
| `internal/provider/run_refs.go` | `abc/runref` |
| `internal/provider/provider.go` job allocation state | `abc/joballoc` |
| `internal/provider/provider_runtime.go` probe tracker | `abc/probe` |
| `internal/openbao/openbao.go` | `impl/openbaoclient` |
| `internal/baopki/baopki.go` | `impl/pkiprovisioner`, `abc/pki` |
| `internal/provider/policy_client.go` | `impl/policyclient`, `abc/policy` |
| `internal/runner/*.go` | `impl/execrunner`, `impl/memoryrunner`, `abc/runner` |
| `internal/denoperm/denoperm.go` | `common/denospec` |
| `internal/{denorun,denojob,denopod,policyengine,policyworkflowpod,policyworkflowrun,trigger}/*.go` | stay in the consumer, re-expressed as `abc/reconcile.Reconciler` deciders |
| `internal/provider/kcpdns/embed.go` | `impl/assets` |
| `api/v1alpha1/types_shared.go` | `common/denospec`, `common/deno` |

## Two tiers

**Unit** — `go test ./...`, no cluster, no binaries, milliseconds. Everything
pure: the queue, the requeue policy, the cache indexers, the status patches,
the runners, the PKI orchestration, the transports against `httptest`.

**Live** — anything that has to talk to kcp. `internal/livekcp` starts a real
kcp in a temp directory, applies an `APIResourceSchema` and an `APIExport`,
creates two consumer workspaces with `APIBinding`s, and waits for them to
bind. Directly against it:

- `impl/kcpstore` — create, get, list, a status merge patch, a finalizer JSON
  patch, the cluster path annotation, delete, and a 404, all through the typed
  resource.
- `factory/controller` — `exportwatch` discovers the real virtual workspace
  URL from the export's endpoint slice, wildcard informers watch it, the
  workqueue delivers keys, and a decider drives two widgets to `Succeeded`.

and the three examples that need a cluster: `examples/controller`,
`examples/admission`, `examples/dns`.

Everything that needs a cluster starts at most one, because kcp's own startup
dominates and it is the same every time:

| | |
|---|---|
| kcp ready to serve (`/readyz` returns 200) | ~10s |
| provision: three workspaces, an export, two bindings | ~0.6s |
| one live test, cluster already up | 0.8 - 2.9s |

`make test-live` and `make examples` run the whole tier against a single
cluster through `scripts/live.sh`, so those ten seconds are paid once. Both
take about 17s in total, against about 43s and 28s if each started its own.

The other thing that made those numbers what they are: `impl/kcpstore` sets
`QPS` and `Burst` on the config it builds, defaulting to 50 and 100. client-go
defaults every REST client to 5 requests a second with a burst of 10, and a
controller that inherits that spends most of its time waiting on its own rate
limiter -- the admission example went from 21.7s to 1.6s when it stopped.
`kcpstore.Options` overrides both, and `kcpstore.Tuned` applies the same
defaults to a config you built yourself.

Run one on its own with `go run ./examples/controller`, or point several at a
cluster that is already up:

```bash
KCP_LIBS_KUBECONFIG=/path/to/admin.kubeconfig KCP_LIBS_SERVER=https://127.0.0.1:6443 \
  go run ./examples/controller
```

kcp runs its own embedded etcd by default, so the only binaries needed are
`kcp` and `kubectl` (override with `KCP_BIN`, `KUBECTL`). That embedded etcd
binds the fixed ports 2379 and 2380, so two of these clusters cannot run at
once; set `KCP_LIBS_ETCD_SERVERS` to a store that is already running (a real
etcd, or a kine on its own port) when you want instances in parallel.

The live tests are opt-in: with `KCP_LIBS_REQUIRE_LIVE=1` they fail instead of
skipping when a binary is missing, which is what CI should set.

## Commands

```bash
make check                    # gofmt -l, go vet, go mod tidy -diff, go test; read-only
make format                   # rewrite what gofmt would change
make test                     # the unit tier
make race
make test-live                # the live tier: a real kcp
make examples                 # run all six examples
go run ./examples/controller  # or just one
```

## License

Unlicense (public domain). See `LICENSE`.
