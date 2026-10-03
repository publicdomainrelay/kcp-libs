package controller

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/abc/driver"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/metrics"
)

type Options struct {
	Config *rest.Config

	Bases []string

	Resources []informerwatch.Resource

	Indexers cache.Indexers

	Reactor informerwatch.Reactor

	Handler driver.Handler

	Policy driver.Policy

	Workers int

	Metrics *metrics.Registry

	Log *slog.Logger

	Now func() time.Time

	IsConflict func(error) bool

	OnEvent func()
}

type Controller struct {
	opts Options

	set *cache.Set

	queue workqueue.TypedRateLimitingInterface[driver.Key]

	reconciles atomic.Uint64

	conflicts atomic.Uint64

	errors atomic.Uint64

	started atomic.Bool

	lastEventNanos atomic.Int64

	reconcileSeconds *metrics.Summary

	queueDepth *metrics.Gauge

	cacheAge *metrics.Gauge
}

func New(opts Options) (*Controller, error) {
	if opts.Config == nil {
		return nil, errors.New("controller: a rest config is required")
	}
	if opts.Handler == nil {
		return nil, errors.New("controller: a handler is required")
	}
	if len(opts.Resources) == 0 {
		return nil, errors.New("controller: at least one resource is required")
	}
	if opts.Policy.Interval <= 0 {
		opts.Policy.Interval = driver.DefaultRequeueAfter
	}
	if opts.Policy.MinTransitionPoll <= 0 {
		opts.Policy.MinTransitionPoll = driver.DefaultMinTransitionPoll
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
	c := &Controller{
		opts:  opts,
		set:   cache.NewSet(),
		queue: workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[driver.Key]()),
	}
	c.reconcileSeconds = opts.Metrics.Summary("reconcile_seconds", "time spent inside a reconcile")
	c.queueDepth = opts.Metrics.Gauge("queue_depth", "work keys waiting in the reconcile queue")
	c.cacheAge = opts.Metrics.Gauge("cache_age_seconds", "seconds since the last informer event")
	return c, nil
}

func DefaultWorkers() int {
	return max(min(runtime.GOMAXPROCS(0), 16), 1)
}

func (c *Controller) Set() *cache.Set {
	return c.set
}

func (c *Controller) Lookup() informerwatch.Lookup {
	return c.set
}

func (c *Controller) Enqueue(kind string, r ref.Ref) {
	c.queue.Add(driver.Key{Kind: kind, Ref: ref.New(r.LogicalCluster, r.Namespace, r.Name)})
}

func (c *Controller) EnqueueAfter(kind string, r ref.Ref, after time.Duration) {
	c.queue.AddAfter(driver.Key{Kind: kind, Ref: ref.New(r.LogicalCluster, r.Namespace, r.Name)}, after)
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

func (c *Controller) Running() bool {
	return c.started.Load()
}

func (c *Controller) Run(ctx context.Context) error {
	c.started.Store(true)
	defer c.started.Store(false)

	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	watchErr := make(chan error, 1)
	go func() {
		watchErr <- informerwatch.Run(watchCtx, informerwatch.Options{
			Config:    c.opts.Config,
			Bases:     c.opts.Bases,
			Resources: c.opts.Resources,
			Indexers:  c.opts.Indexers,
			Set:       c.set,
			Reactor:   c.opts.Reactor,
			Enqueue:   c.Enqueue,
			OnEvent:   c.onEvent,
		})
	}()

	var wg sync.WaitGroup
	for i := 0; i < c.opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.worker(ctx)
		}()
	}

	var err error
	select {
	case <-ctx.Done():
	case err = <-watchErr:
		cancelWatch()
	}
	c.queue.ShutDown()
	wg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (c *Controller) onEvent() {
	c.RecordEvent()
	c.cacheAge.Set(0)
	if c.opts.OnEvent != nil {
		c.opts.OnEvent()
	}
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
			c.reconcileSeconds.Observe(c.opts.Now().Sub(start))
			c.queueDepth.SetInt(int64(c.queue.Len()))
			if err != nil {
				if c.opts.IsConflict(err) {
					c.conflicts.Add(1)
					c.queue.Forget(key)
					c.queue.AddAfter(key, c.opts.Policy.ConflictAfter(key))
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
