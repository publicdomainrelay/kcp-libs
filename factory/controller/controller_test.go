package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/abc/reconcile"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
)

func newTestController(t *testing.T, handler reconcile.Handler, policy reconcile.Policy) *Controller {
	t.Helper()
	controller, err := New(Options{
		Config:  &rest.Config{Host: "https://kcp.invalid"},
		Sources: []informerwatch.Source{{Base: "https://kcp.invalid", Resources: []informerwatch.Resource{{Kind: "denorun", GVR: schema.GroupVersionResource{Group: "deno.computer", Version: "v1alpha1", Resource: "denoruns"}}}}},
		Handler: handler,
		Policy:  policy,
		Now:     time.Now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return controller
}

func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestWorkerStopsOnATerminalKey(t *testing.T) {
	processed := make(chan reconcile.Key, 4)
	controller := newTestController(t, reconcile.HandlerFunc(func(_ context.Context, key reconcile.Key) (time.Duration, bool, error) {
		processed <- key
		return 0, true, nil
	}), reconcile.Policy{Interval: time.Hour})

	go controller.worker(context.Background())
	defer controller.queue.ShutDown()
	controller.Enqueue("denorun", ref.New("root:alice", "default", "run-1"))

	select {
	case key := <-processed:
		if key.Kind != "denorun" || key.Ref.Name != "run-1" {
			t.Fatalf("key = %+v", key)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the handler was never called")
	}
	waitFor(t, "the queue to drain", func() bool { return controller.QueueDepth() == 0 })
	time.Sleep(20 * time.Millisecond)
	if controller.QueueDepth() != 0 {
		t.Fatal("a terminal key with no requeue must not be rescheduled")
	}
	if controller.Reconciles() != 1 {
		t.Fatalf("reconciles = %d", controller.Reconciles())
	}
}

func TestWorkerRequeuesANonTerminalKey(t *testing.T) {
	processed := make(chan struct{}, 8)
	controller := newTestController(t, reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		processed <- struct{}{}
		return time.Millisecond, false, nil
	}), reconcile.Policy{Interval: time.Millisecond})

	go controller.worker(context.Background())
	defer controller.queue.ShutDown()
	controller.Enqueue("denorun", ref.New("root:alice", "default", "run-1"))

	for i := range 3 {
		select {
		case <-processed:
		case <-time.After(5 * time.Second):
			t.Fatalf("the handler ran %d times, want at least 3", i)
		}
	}
}

func TestWorkerRequeuesAConflictAtTheClamp(t *testing.T) {
	controller := newTestController(t, reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		return 0, false, apierrors.NewConflict(schema.GroupResource{Resource: "denoruns"}, "run-1", errors.New("stale"))
	}), reconcile.Policy{
		Interval:          time.Hour,
		MinTransitionPoll: time.Millisecond,
		ClampKinds:        map[string]bool{"denorun": true},
	})

	go controller.worker(context.Background())
	defer controller.queue.ShutDown()
	controller.Enqueue("denorun", ref.New("root:alice", "default", "run-1"))

	waitFor(t, "the conflict to be counted", func() bool { return controller.Conflicts() >= 1 })
	if controller.Errors() != 0 {
		t.Fatal("a conflict is not an error")
	}
	waitFor(t, "the key to be requeued", func() bool { return controller.QueueDepth() == 1 })
}

func TestWorkerRateLimitsAnError(t *testing.T) {
	controller := newTestController(t, reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		return 0, false, errors.New("boom")
	}), reconcile.Policy{Interval: time.Hour})

	go controller.worker(context.Background())
	defer controller.queue.ShutDown()
	controller.Enqueue("denorun", ref.New("root:alice", "default", "run-1"))

	waitFor(t, "the error to be counted", func() bool { return controller.Errors() >= 1 })
}

func TestNewRequiresItsWiring(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("a rest config is required")
	}
	if _, err := New(Options{Config: &rest.Config{}}); err == nil {
		t.Fatal("a handler is required")
	}
	if _, err := New(Options{Config: &rest.Config{}, Handler: reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		return 0, true, nil
	})}); err == nil {
		t.Fatal("at least one resource is required")
	}
}

func TestDefaultWorkersIsBounded(t *testing.T) {
	workers := DefaultWorkers()
	if workers < 1 || workers > 16 {
		t.Fatalf("workers = %d", workers)
	}
}

func TestCacheAgeMeasuresTheInformerNotTheReconciler(t *testing.T) {
	now := time.Unix(1000, 0)
	controller := newTestController(t, reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		return 0, true, nil
	}), reconcile.Policy{Interval: time.Hour})
	controller.opts.Now = func() time.Time { return now }

	if age := controller.CacheAge(); age != 0 {
		t.Fatalf("before any event the age is unknown, got %v", age)
	}
	controller.onEvent()
	now = now.Add(30 * time.Second)
	if age := controller.CacheAge(); age != 30*time.Second {
		t.Fatalf("age = %v, want 30s since the informer's event", age)
	}

	processed := make(chan struct{}, 1)
	controller.opts.Handler = reconcile.HandlerFunc(func(context.Context, reconcile.Key) (time.Duration, bool, error) {
		processed <- struct{}{}
		return 0, true, nil
	})
	go controller.worker(context.Background())
	defer controller.queue.ShutDown()
	controller.Enqueue("widget", ref.New("root:alice", "default", "one"))
	select {
	case <-processed:
	case <-time.After(5 * time.Second):
		t.Fatal("the handler never ran")
	}
	if age := controller.CacheAge(); age != 30*time.Second {
		t.Fatalf("a reconcile must not reset the cache age, got %v", age)
	}
}
