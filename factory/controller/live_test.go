package controller_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/reconcile"
	"github.com/publicdomainrelay/kcp-libs/common/logging"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
	"github.com/publicdomainrelay/kcp-libs/factory/controller"
	"github.com/publicdomainrelay/kcp-libs/impl/exportwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

type probeSpec struct {
	Steps int32 `json:"steps,omitempty"`
}

type probeStatus struct {
	Phase string `json:"phase,omitempty"`

	Observed int32 `json:"observed,omitempty"`
}

type probe = livekcp.Object[probeSpec, probeStatus]

func TestLiveControllerAgainstRealKCP(t *testing.T) {
	livekcp.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cluster, err := livekcp.Start(ctx)
	if err != nil {
		t.Fatalf("start kcp: %v", err)
	}
	defer cluster.Stop()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.Server, RestConfig: cluster.Config})
	if err != nil {
		t.Fatal(err)
	}
	probes := kcpstore.Of[probe](store, livekcp.WidgetGVR)

	endpoints, err := exportwatch.Await(ctx, exportwatch.Options{
		Config:            store.Config(),
		Host:              cluster.Server,
		ProviderWorkspace: cluster.ProviderCluster,
		Exports:           []string{livekcp.Export, livekcp.SecondExport},
		Log:               logging.Discard(),
	})
	if err != nil {
		t.Fatalf("discover the virtual workspaces: %v", err)
	}
	probeBase := exportwatch.Paths(endpoints, livekcp.Export)
	gadgetBase := exportwatch.Paths(endpoints, livekcp.SecondExport)
	if len(probeBase) == 0 || len(gadgetBase) == 0 {
		t.Fatalf("kcp published %d probe and %d gadget endpoints", len(probeBase), len(gadgetBase))
	}
	t.Logf("the probes and the gadgets are served by different exports: %s and %s", probeBase[0], gadgetBase[0])

	gadgets := kcpstore.Of[probe](store, livekcp.GadgetGVR)
	resourceFor := func(kind string) *kcpstore.Resource[probe] {
		if kind == "gadget" {
			return gadgets
		}
		return probes
	}

	handler := reconcile.HandlerFunc(func(ctx context.Context, key reconcile.Key) (time.Duration, bool, error) {
		resource := resourceFor(key.Kind)
		obj, err := resource.Get(ctx, key.Ref)
		if err != nil {
			if kcpstore.IsNotFound(err) {
				return 0, true, nil
			}
			return 0, false, err
		}
		phase := obj.Status.Phase
		observed := obj.Status.Observed
		switch phase {
		case "":
			phase = "Pending"
		case "Pending":
			phase = "Running"
		case "Running":
			observed++
			if observed >= obj.Spec.Steps {
				phase = "Succeeded"
			}
		default:
			return 0, true, nil
		}
		patch, err := statuspatch.Merge(map[string]any{"phase": phase, "observed": observed})
		if err != nil {
			return 0, false, err
		}
		if err := resource.PatchStatus(ctx, key.Ref.WithResourceVersion(obj.Metadata.ResourceVersion), patch); err != nil {
			return 0, false, err
		}
		return 10 * time.Millisecond, phase == "Succeeded", nil
	})

	ctl, err := controller.New(controller.Options{
		Config: store.Config(),
		Sources: []informerwatch.Source{
			{Base: probeBase[0], Resources: []informerwatch.Resource{{Kind: "probe", GVR: livekcp.WidgetGVR}}},
			{Base: gadgetBase[0], Resources: []informerwatch.Resource{{Kind: "gadget", GVR: livekcp.GadgetGVR}}},
		},
		Handler: handler,
		Policy: reconcile.Policy{
			Interval:          200 * time.Millisecond,
			MinTransitionPoll: 10 * time.Millisecond,
			ClampKinds:        map[string]bool{"probe": true},
		},
		Log: logging.New(logging.Options{Service: "live", Writer: testWriter{t}, Level: slog.LevelError}),
	})
	if err != nil {
		t.Fatal(err)
	}
	runCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ctl.Run(runCtx)
	}()
	defer func() {
		stop()
		<-done
	}()

	seed(ctx, t, probes, cluster.ConsumerCluster, "alpha", 2)
	waitPhase(ctx, t, probes, cluster.ConsumerCluster, "alpha", "Succeeded")
	seed(ctx, t, probes, cluster.ConsumerCluster, "beta", 1)
	waitPhase(ctx, t, probes, cluster.ConsumerCluster, "beta", "Succeeded")
	seedGadget(ctx, t, gadgets, cluster.ConsumerCluster, "gear", 1)
	waitGadget(ctx, t, gadgets, cluster.ConsumerCluster, "gear", "Succeeded")

	if ctl.Reconciles() == 0 {
		t.Fatal("the controller reconciled nothing")
	}
	t.Logf("reconciles %d, queue depth %d, cache age %s", ctl.Reconciles(), ctl.QueueDepth(), ctl.CacheAge().Round(time.Millisecond))
}

func seed(ctx context.Context, t *testing.T, probes *kcpstore.Resource[probe], cluster, name string, steps int32) {
	t.Helper()
	obj := livekcp.NewObject[probeSpec, probeStatus]("default", name)
	obj.Spec.Steps = steps
	if err := livekcp.Seed(ctx, cluster, probes, "default", name, obj); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
}

func seedGadget(ctx context.Context, t *testing.T, gadgets *kcpstore.Resource[probe], cluster, name string, steps int32) {
	t.Helper()
	obj := livekcp.NewObject[probeSpec, probeStatus]("default", name)
	obj.Kind = livekcp.SecondKind
	obj.Spec.Steps = steps
	if err := livekcp.Seed(ctx, cluster, gadgets, "default", name, obj); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
}

func waitGadget(ctx context.Context, t *testing.T, gadgets *kcpstore.Resource[probe], cluster, name, want string) {
	t.Helper()
	if _, err := livekcp.WaitFor(ctx, cluster, gadgets, "default", name, func(obj probe) bool {
		return obj.Status.Phase == want
	}); err != nil {
		t.Fatalf("gadget %v", err)
	}
}

func waitPhase(ctx context.Context, t *testing.T, probes *kcpstore.Resource[probe], cluster, name, want string) {
	t.Helper()
	if _, err := livekcp.WaitFor(ctx, cluster, probes, "default", name, func(obj probe) bool {
		return obj.Status.Phase == want
	}); err != nil {
		t.Fatalf("%v", err)
	}
}

type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(body []byte) (int, error) {
	w.t.Log(string(body))
	return len(body), nil
}
