# Examples

Six runnable programs. Each one is a whole story about one part of the library:
with a real kcp running, a CRD is created, a controller or a loop acts on it,
and the result is printed. Read them in this order and you have seen everything
the module does.

```bash
go run ./examples/controller    # the reconcile loop, over a real informer
go run ./examples/admission     # the queue, and the cap it enforces
go run ./examples/workloads     # processes: argv, outputs, probes, ttl
go run ./examples/pki           # one certificate authority per namespace
go run ./examples/policy        # submitting a workflow and reading the verdict
go run ./examples/dns           # cluster-local names and per-workspace tokens
```

`controller`, `admission` and `dns` need `kcp` and `kubectl` on PATH
(`KCP_BIN`, `KUBECTL` override). Each starts a real kcp -- with its own
embedded etcd -- in a temp directory and drives it. `make examples` starts one
cluster for all six, because kcp's own startup is about ten seconds and is the
same every time; point `KCP_LIBS_KUBECONFIG` at an existing cluster to do the
same by hand.

Watching two exports at once, each with its own resource set, is not an
example: it is `factory/controller`'s live test, which reconciles a `Widget`
from one export and a `Gadget` from another because that is the only way to
prove the driver keeps them apart.

The three that need a cluster share one permissive object, a `Widget`:
`livekcp.Object[Spec, Status]` is the envelope, so an example writes only the
fields its story needs.

The other three need no cluster: `workloads` runs real processes, and `pki` and
`policy` talk to a fake vault and a fake policy engine -- separate products
with nothing to install here. Both clients are real: `openbaoclient` wraps the
official `github.com/openbao/openbao/api/v2`, and the fake vault answers the
same wire protocol that client speaks.

`go test ./examples/...` asserts the lines each example prints. The three that
need kcp skip unless `KCP_LIBS_REQUIRE_LIVE=1` is set; `scripts/live.sh`, which
both `make test-live` and `make examples` go through, exports it. Those two
targets run the examples rather than their tests, so to assert the kcp-backed
three run `KCP_LIBS_REQUIRE_LIVE=1 go test ./examples/...`; the other three
always run. `make examples` starts one
cluster for all six and takes about 17s; the three that need it take 1 - 2s
each once it is up.

## The shape every example has

```go
func Run(ctx context.Context, out io.Writer) error
func main() { Run(context.Background(), os.Stdout) }
```

`Run` is the example and the test fixture at once. `main` prints the story;
`main_test.go` asserts the exact lines it prints. Output is the observable
behaviour, so a broken library breaks a line of prose.

`main.go` is the whole story. When an example needs a service to talk to, the
fake for it gets its own file: `pki/vault.go` and `pki/certs.go` are a vault
that issues real certificates, `policy/engine.go` is an engine that finishes a
task after a few polls.

## The cluster

`internal/livekcp` starts kcp in a temp directory, applies APIResourceSchemas
and APIExports for two example kinds (`Widget` and `Gadget`, in
`example.computer/v1alpha1`, each with permissive `spec` and `status`),
creates a provider workspace and two consumer workspaces bound to both
exports, and waits for the bindings. Two exports rather than one because a
resource may only be watched against the export that serves it, and the
controller test proves the driver can tell them apart.

That is why `examples/controller` shows a real watch-driven controller: the
virtual workspace URL comes from the export's endpoint slice, the wildcard
informers watch it, and the status patches land on a real API server.

The harness is not importable from the library layers: `internal/boundaries`
fails the test tier if anything outside `examples/` and tests reaches into
`internal/`.

## When you would reach for each package

| Package | Reach for it when | Shown in |
|---|---|---|
| `abc/reconcile` | you are writing a decider, or you need the requeue rules; `Bridge` turns one into the other and `Patch` turns a result into the merge patch | `controller` `Bridge`, `Patch`, `Policy` |
| `abc/cache` | you want the informer's objects by index instead of a list, and you own the set | `controller` `cache.NewSet`, `IndexersFor`, `ByIndex` |
| `factory/controller` | you want informers plus a workqueue plus a worker pool, wired | `controller` `controller.New` |
| `impl/kcpstore` | you need to read or write a CRD on kcp, typed or raw | `controller`, `admission`, `dns` |
| `impl/exportwatch` | you need the APIExport virtual workspace URL at startup | `controller` `exportwatch.Await` |
| `impl/informerwatch` | you want the informers without the controller around them | `controller` (inside `factory/controller`) |
| `impl/metrics` | you want the queue depth, cache age and reconcile timings | `controller` `metrics.New` |
| `abc/queue` | a parent caps how many children run at once | `admission` `Admit` |
| `factory/admission` | the same, composed with your store and your wake function | `admission` `admission.New` |
| `abc/runref` | a stateless decider could start the same workload twice | `admission` `runref.AlreadyStarted` |
| `abc/runner` | you are starting processes and probing them | `workloads` |
| `impl/execrunner` | the real one: `os/exec`, run directories, process groups | `workloads` `execrunner.NewPod` |
| `impl/memoryrunner` | the same interface with no processes, for unit tests | `workloads` `memoryrunner.NewPod` |
| `impl/assets` | a shim or a probe has to exist on disk before a run starts | `dns` `assets.DNSSet`, `workloads` `assets.Set` |
| `common/denospec` | permissions become `deno run` arguments | `workloads` `denospec.Args` |
| `abc/probe` | a liveness probe counts failures before a restart | `workloads` `probe.NewTracker` |
| `common/ttl` | retention after a run finishes, or an active deadline | `workloads` `ttl.Expired` |
| `abc/joballoc` | the names a job created are not yet visible through the API | `workloads` `joballoc.New` |
| `abc/pki` | you are issuing certificates and want the port, not OpenBao | `pki` |
| `impl/openbaoclient` | the OpenBao HTTP API, typed, over the official `api/v2` client | `pki` `openbaoclient.New` |
| `impl/pkiprovisioner` | one intermediate per namespace, root in the root namespace | `pki` `EnsureAuthority` |
| `abc/policy` | you submit a workflow and poll for its verdict | `policy` |
| `impl/policyclient` | the gha-lite engine's HTTP API, verdict unwrapping included | `policy` `policyclient.New` |
| `common/outputs` | a JSON output map has to become flat strings | `impl/policyclient` (see `policy`) |
| `factory/servicenames` | a workload resolves a peer by name, and tokens are per workspace | `dns` |
| `common/kcp` | a workspace path has to become DNS labels, or back | `dns` `ServiceLabels` |
| `common/ref` | you need the identity of an object: cluster, namespace, name | `controller`, `admission`, `dns`, `workloads` |
| `common/clientlimit` | you are building a rest config and want more than client-go's 5 requests a second | `impl/kcpstore`, `impl/exportwatch`, `impl/informerwatch` |
| `common/expiringmap` | you need a ttl-bounded map | `abc/queue` `Leases`, `abc/runref`, `abc/joballoc`, `impl/pkiprovisioner` |
| `common/statuspatch` | the status write is a merge patch, or a finalizer patch | `controller`, `admission` |
| `common/condition` | a status carries `metav1.Condition` and you edit one | `controller` `decide` |
| `common/denocomputer` | you are working in the `deno.computer` vocabulary | `controller` phases and condition |
| `common/logging` | a JSON slog logger, or one that throws output away | `controller` `logging.New` |
| `abc/store` | the interfaces `impl/kcpstore` implements, and `Unchanged` | `controller` |
| `internal/livekcp` | you want the examples and live tests to have a real kcp | `controller`, `admission`, `dns` |

## What each example is about

**`controller`**  - the whole loop. A `Widget` is created in a workspace, the
informer sees it, the workqueue delivers the key, the decider says what phase
it should be in, and the handler patches the status. It shows the difference
between the list path (two widgets seeded before the controller starts) and the
watch path (a third created after), and reads the informer cache by index to
count the widgets in a group.

**`admission`**  - the queue. A batch allows two items at once and five are
created, so three wait. It prints the `AtCapacity` message one of them
carries, reports the peak number of items observed running against the cap of
two, and shows the lease the admission takes so two passes cannot both see a
free slot. The duplicate-start guard is exercised both ways: five copies that
predate their own write are refused, and ten legitimate starts -- retries and
recreated objects -- are allowed through.

**`workloads`**  - processes. Permissions become argv, a stand-in runtime is
materialised, the process runs, its `result.json` becomes outputs, a probe
runs in the run directory, a liveness tracker asks for a restart, retention
decides on a delete, and the same `PodRunner` interface answers without a
process at all.

**`pki`**  - certificates. A fake vault implements the OpenBao API, the
provisioner generates a root, gives a namespace an intermediate signed by it,
writes the role, and issues a leaf for the workload's service name. The
authority is cached (a second leaf costs one call), and deleting the namespace
makes the next provisioning do the work again.

**`policy`**  - the policy engine. A workflow is submitted, polled to a
terminal state, and the verdict comes back flattened into outputs. A failed
workflow reports why, and an unreachable engine leaves the task running rather
than failing it.

**`dns`**  - names. Workloads in two workspaces advertise addresses in their own
spec, and the example builds the FQDN table from them, injects a workload's own
name before it can be observed, and mints one token per workspace through a
real `TokenRequest`. The FQDNs carry the workspace *path*, read from the
cluster's own `kcp.io/path` annotation, so they read `pds.default.consumer`
rather than an opaque workspace id.
