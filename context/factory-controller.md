# Context: factory-controller

Repository: `kcp-libs`

This context is the controller runtime of kcp-libs: the generic work-queue driver that turns watched kcp resources into reconcile invocations. It exists so the deno workload packages and downstream consumers can share one tested scheduling core — bounded-concurrency workers, rate limited retries, optimistic-concurrency conflict handling, informer-event freshness tracking and prometheus metrics — instead of each rebuilding it. It names no workload runtime; the deno vocabulary lives in the consumers, and the controller only knows kinds, refs and a reconcile.Handler.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
