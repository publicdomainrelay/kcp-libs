package controller

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/abc/reconcile"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/metrics"
)

type Options struct {
	Config *rest.Config

	Sources []informerwatch.Source

	Indexers cache.Indexers

	Set *cache.Set

	Reactor informerwatch.Reactor

	Handler reconcile.Handler

	Policy reconcile.Policy

	Workers int

	Metrics *metrics.Registry

	Log *slog.Logger

	Now func() time.Time

	IsConflict func(error) bool
}

type Controller struct {
	opts Options

	set *cache.Set

	queue workqueue.TypedRateLimitingInterface[reconcile.Key]

	reconciles atomic.Uint64

	conflicts atomic.Uint64

	errors atomic.Uint64

	lastEventNanos atomic.Int64

	reconcileSeconds prometheus.Summary

	conflictsTotal prometheus.Counter

	queueDepth prometheus.GaugeFunc

	cacheAge prometheus.GaugeFunc
}

func New(opts Options) (*Controller, error) {
	if opts.Config == nil {
		return nil, errors.New("controller: a rest config is required")
	}
	if opts.Handler == nil {
		return nil, errors.New("controller: a handler is required")
	}
	if len(opts.Sources) == 0 {
		return nil, errors.New("controller: at least one source is required")
	}
	if opts.Policy.Interval <= 0 {
		opts.Policy.Interval = reconcile.DefaultRequeueAfter
	}
	if opts.Policy.MinTransitionPoll <= 0 {
		opts.Policy.MinTransitionPoll = reconcile.DefaultMinTransitionPoll
	}
	if opts.Workers <= 0 {
		opts.Workers = DefaultWorkers()
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	if opts.IsConflict == nil {
		opts.IsConflict = apierrors.IsConflict
	}
	if opts.Metrics == nil {
		opts.Metrics = metrics.New("controller")
	}
	if opts.Set == nil {
		opts.Set = cache.NewSet()
	}
	c := &Controller{
		opts:  opts,
		set:   opts.Set,
		queue: workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Key]()),
	}
	c.registerMetrics()
	return c, nil
}

func (c *Controller) registerMetrics() {
	opts := c.opts
	c.reconcileSeconds = opts.Metrics.Summary("reconcile_seconds", "time spent inside a reconcile")
	c.conflictsTotal = opts.Metrics.Counter("conflicts_total", "writes rejected by optimistic concurrency")

	c.queueDepth = opts.Metrics.GaugeFunc("queue_depth", "work keys that are ready to reconcile", func() float64 {
		return float64(c.queue.Len())
	})
	c.cacheAge = opts.Metrics.GaugeFunc("cache_age_seconds", "seconds since the last informer event, NaN before the first one", func() float64 {
		if c.lastEventNanos.Load() == 0 {
			return math.NaN()
		}
		return c.CacheAge().Seconds()
	})
}

func DefaultWorkers() int {
	return max(min(runtime.GOMAXPROCS(0), 16), 1)
}

func (c *Controller) Enqueue(kind string, r ref.Ref) {
	c.queue.Add(reconcile.Key{Kind: kind, Ref: ref.New(r.LogicalCluster, r.Namespace, r.Name)})
}

func (c *Controller) QueueDepth() int {
	return c.queue.Len()
}

func (c *Controller) Reconciles() uint64 {
	return c.reconciles.Load()
}

func (c *Controller) Conflicts() uint64 {
	return c.conflicts.Load()
}

func (c *Controller) Errors() uint64 {
	return c.errors.Load()
}

func (c *Controller) Run(ctx context.Context) error {

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	watchErr := make(chan error, 1)
	go func() {
		watchErr <- informerwatch.Run(runCtx, informerwatch.Options{
			Config:   c.opts.Config,
			Sources:  c.opts.Sources,
			Indexers: c.opts.Indexers,
			Set:      c.set,
			Reactor:  c.opts.Reactor,
			Enqueue:  c.Enqueue,
			OnEvent:  c.onEvent,
		})
	}()

	var wg sync.WaitGroup
	for i := 0; i < c.opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.worker(runCtx)
		}()
	}

	var err error
	select {
	case <-ctx.Done():
	case err = <-watchErr:
		cancel()
	}
	c.queue.ShutDown()
	wg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (c *Controller) onEvent() {
	c.recordEvent()
}

func (c *Controller) worker(ctx context.Context) {
	for {
		key, shutdown := c.queue.Get()
		if shutdown {
			return
		}
		func() {
			defer c.queue.Done(key)
			c.reconciles.Add(1)
			start := c.opts.Now()
			after, terminal, err := c.opts.Handler.Process(ctx, key)
			c.reconcileSeconds.Observe(c.opts.Now().Sub(start).Seconds())
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if c.opts.IsConflict(err) {
					c.conflicts.Add(1)
					c.conflictsTotal.Inc()
					c.queue.Forget(key)
					c.queue.AddAfter(key, c.opts.Policy.ConflictAfter())
					return
				}
				c.errors.Add(1)
				c.opts.Log.Error("reconcile failed",
					"kind", key.Kind, "workspace", key.Ref.LogicalCluster, "name", key.Ref.Name, "err", err)
				c.queue.AddRateLimited(key)
				return
			}
			c.queue.Forget(key)
			delay, requeue := c.opts.Policy.Next(key, after, terminal)
			if !requeue {
				return
			}
			c.queue.AddAfter(key, delay)
		}()
	}
}
