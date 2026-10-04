# Context: factory-controller

Repository: `kcp-libs`

This context is the controller runtime of kcp-libs: the generic work-queue driver that turns watched kcp resources into reconcile invocations. It exists so the deno workload packages and downstream consumers can share one tested scheduling core — bounded-concurrency workers, rate limited retries, optimistic-concurrency conflict handling, informer-event freshness tracking and prometheus metrics — instead of each rebuilding it. It names no workload runtime; the deno vocabulary lives in the consumers, and the controller only knows kinds, refs and a reconcile.Handler.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: factory/controller/controller.go
  kind: struct
  name: Controller
  signature: type Controller struct
- file: factory/controller/events.go
  kind: method
  name: Controller.CacheAge
  signature: func (c *Controller) CacheAge() time.Duration
- file: factory/controller/controller.go
  kind: method
  name: Controller.Conflicts
  signature: func (c *Controller) Conflicts() uint64
- file: factory/controller/controller.go
  kind: method
  name: Controller.Enqueue
  signature: func (c *Controller) Enqueue(kind string, r ref.Ref)
- file: factory/controller/controller.go
  kind: method
  name: Controller.Errors
  signature: func (c *Controller) Errors() uint64
- file: factory/controller/controller.go
  kind: method
  name: Controller.QueueDepth
  signature: func (c *Controller) QueueDepth() int
- file: factory/controller/controller.go
  kind: method
  name: Controller.Reconciles
  signature: func (c *Controller) Reconciles() uint64
- file: factory/controller/controller.go
  kind: method
  name: Controller.Run
  signature: func (c *Controller) Run(ctx context.Context) error
- file: factory/controller/controller.go
  kind: function
  name: DefaultWorkers
  signature: func DefaultWorkers() int
- file: factory/controller/controller.go
  kind: function
  name: New
  signature: func New(opts Options) (*Controller, error)
- file: factory/controller/controller.go
  kind: struct
  name: Options
  signature: type Options struct
requirements:
- codeRefs:
  - file:factory/controller/controller.go
  - function:9074c79f1f56d6b08908fad3482eacf7
  id: r.default-workers-clamped
  level: MUST
  text: DefaultWorkers returns runtime.GOMAXPROCS(0) clamped to the range 1 through
    16.
- codeRefs:
  - file:factory/controller/controller.go
  - method:e5158999a5682e7a4c200ae884ceaaa7
  id: r.enqueue-key
  level: MUST
  text: Enqueue builds a reconcile.Key from the kind and a ref.New of the ref.Ref's
    LogicalCluster, Namespace and Name, and adds it to the rate limiting workqueue.
- codeRefs:
  - file:factory/controller/events.go
  - method:e95e98357d5f3058033476c49387e673
  id: r.event-freshness
  level: SHOULD
  text: The OnEvent hook wired into informerwatch records the current Now timestamp
    so that CacheAge and cache_age_seconds reflect informer liveness.
- codeRefs:
  - file:factory/controller/controller_test.go
  - file:factory/controller/live_test.go
  id: r.live-verification
  level: SHOULD
  text: The controller is exercised against a real kcp cluster by TestLiveControllerAgainstRealKCP
    using kcpstore and exportwatch wiring.
- codeRefs:
  - file:factory/controller/controller.go
  - method:7b04bad45e95e9143fa55f0f8586bbdb
  - method:e95e98357d5f3058033476c49387e673
  - struct:6973c4a29350d7c92098bf48291ceafd
  id: r.metrics-families
  level: MUST
  text: New registers reconcile_seconds (summary), conflicts_total (counter), queue_depth
    (gauge function over the queue length) and cache_age_seconds (gauge function over
    CacheAge, reporting NaN before the first event) on the chosen metrics registry.
- codeRefs:
  - file:factory/controller/controller.go
  - function:9074c79f1f56d6b08908fad3482eacf7
  - function:e71a19275883cd3b3727e16c601067ec
  id: r.new-defaults-options
  level: MUST
  text: 'New fills absent options: Policy.Interval from reconcile.DefaultRequeueAfter,
    Policy.MinTransitionPoll from reconcile.DefaultMinTransitionPoll, Workers from
    DefaultWorkers, Now from time.Now, Log from slog.Default, IsConflict from apierrors.IsConflict,
    Metrics from metrics.New("controller") and Set from cache.NewSet.'
- codeRefs:
  - file:factory/controller/controller.go
  - function:e71a19275883cd3b3727e16c601067ec
  id: r.new-validates-wiring
  level: MUST
  text: 'New returns an error when Config is nil ("controller: a rest config is required"),
    when Handler is nil ("controller: a handler is required"), or when Sources is
    empty ("controller: at least one source is required"), and returns a usable *Controller
    otherwise.'
- codeRefs:
  - file:factory/controller/events.go
  - method:1a8c9e2899cff2770239be37b7e2356e
  - method:627af13a96e55d3834be1684b1f3a926
  - method:ba8012ce4989a1a73fa61d029f34146c
  - method:e3736f862df6f3ade0940b347f611edb
  - method:e95e98357d5f3058033476c49387e673
  id: r.observability-readers
  level: MUST
  text: QueueDepth returns the workqueue length, Reconciles, Conflicts and Errors
    return the respective atomic counters, and CacheAge returns zero before the first
    informer event and otherwise the elapsed time from the last event to Now.
- codeRefs:
  - file:factory/controller/controller.go
  - method:7b04bad45e95e9143fa55f0f8586bbdb
  id: r.run-supervises
  level: MUST
  text: Run starts informerwatch.Run in a goroutine with Config, Sources, Indexers,
    Set, Reactor, Enqueue and OnEvent wired from the options, starts one worker per
    Workers, then waits on ctx.Done() or the watch error channel; it shuts the queue
    down, waits for all workers, and returns the watch error unless it is context.Canceled.
- codeRefs:
  - file:factory/controller/controller.go
  - method:7b04bad45e95e9143fa55f0f8586bbdb
  - struct:6973c4a29350d7c92098bf48291ceafd
  id: r.worker-conflict-path
  level: MUST
  text: A worker that receives an error matching opts.IsConflict increments the conflicts
    counter and the conflicts_total metric, forgets the key, re-queues it after Policy.ConflictAfter(),
    and does not increment the errors counter or log.
- codeRefs:
  - file:factory/controller/controller.go
  - method:7b04bad45e95e9143fa55f0f8586bbdb
  id: r.worker-error-path
  level: MUST
  text: A worker that receives a non-conflict error increments the errors counter,
    logs "reconcile failed" with the kind, workspace, name and error, and re-queues
    the key rate limited; the handler call is timed into the reconcile_seconds summary.
- codeRefs:
  - file:factory/controller/controller.go
  - method:7b04bad45e95e9143fa55f0f8586bbdb
  id: r.worker-success-path
  level: MUST
  text: A worker that handles a key without error forgets it, asks Policy.Next(key,
    after, terminal) for a delay and requeue decision, and re-queues the key after
    that delay only when requeue is true.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:factory/controller/controller.go` file controller.go (factory/controller/controller.go)
- `file:factory/controller/controller_test.go` file controller_test.go (factory/controller/controller_test.go)
- `file:factory/controller/events.go` file events.go (factory/controller/events.go)
- `file:factory/controller/live_test.go` file live_test.go (factory/controller/live_test.go)
- `function:9074c79f1f56d6b08908fad3482eacf7` function DefaultWorkers (factory/controller/controller.go)
- `function:e71a19275883cd3b3727e16c601067ec` function New (factory/controller/controller.go)
- `method:1a8c9e2899cff2770239be37b7e2356e` method Controller.Conflicts (factory/controller/controller.go)
- `method:627af13a96e55d3834be1684b1f3a926` method Controller.QueueDepth (factory/controller/controller.go)
- `method:7b04bad45e95e9143fa55f0f8586bbdb` method Controller.Run (factory/controller/controller.go)
- `method:ba8012ce4989a1a73fa61d029f34146c` method Controller.Errors (factory/controller/controller.go)
- `method:e3736f862df6f3ade0940b347f611edb` method Controller.Reconciles (factory/controller/controller.go)
- `method:e5158999a5682e7a4c200ae884ceaaa7` method Controller.Enqueue (factory/controller/controller.go)
- `method:e95e98357d5f3058033476c49387e673` method Controller.CacheAge (factory/controller/events.go)
- `struct:6973c4a29350d7c92098bf48291ceafd` struct Controller (factory/controller/controller.go)
- `struct:d2226a6c30fcfabce22cf7b5d8a0d19a` struct Options (factory/controller/controller.go)
<!-- SPECD_MANAGED_END -->
