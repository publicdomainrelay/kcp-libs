package controller_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/publicdomainrelay/kcp-libs/abc/driver"
	"github.com/publicdomainrelay/kcp-libs/common/logging"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
	"github.com/publicdomainrelay/kcp-libs/factory/controller"
	"github.com/publicdomainrelay/kcp-libs/impl/exportwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

var liveProbes = schema.GroupVersionResource{Group: livekcp.Group, Version: livekcp.Version, Resource: livekcp.Resource}

type liveProbe struct {
	APIVersion string `json:"apiVersion,omitempty"`

	Kind string `json:"kind,omitempty"`

	Metadata struct {
		Name string `json:"name"`

		Namespace string `json:"namespace"`

		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`

	Spec struct {
		Steps int32 `json:"steps"`
	} `json:"spec"`

	Status struct {
		Phase string `json:"phase,omitempty"`

		Observed int32 `json:"observed,omitempty"`
	} `json:"status"`
}

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
	probes := kcpstore.Of[liveProbe](store, liveProbes)

	endpoints, err := exportwatch.Await(ctx, exportwatch.Options{
		Config:            store.Config(),
		Host:              cluster.Server,
		ProviderWorkspace: cluster.Provider,
		Exports:           []string{livekcp.Export},
		Log:               logging.Discard(),
	})
	if err != nil {
		t.Fatalf("discover the virtual workspace: %v", err)
	}
	bases := exportwatch.Paths(endpoints, livekcp.Export)
	if len(bases) == 0 {
		t.Fatal("kcp published no virtual workspace URL")
	}
	t.Logf("virtual workspace: %s", bases[0])

	handler := driver.HandlerFunc(func(ctx context.Context, key driver.Key) (time.Duration, bool, error) {
		obj, err := probes.Get(ctx, key.Ref)
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
		if err := probes.PatchStatus(ctx, key.Ref.WithResourceVersion(obj.Metadata.ResourceVersion), patch); err != nil {
			return 0, false, err
		}
		return 10 * time.Millisecond, phase == "Succeeded", nil
	})

	ctl, err := controller.New(controller.Options{
		Config:    store.Config(),
		Bases:     bases,
		Resources: []informerwatch.Resource{{Kind: "probe", GVR: liveProbes}},
		Handler:   handler,
		Policy: driver.Policy{
			Interval:          200 * time.Millisecond,
			MinTransitionPoll: 10 * time.Millisecond,
			ClampKinds:        map[string]bool{"probe": true},
		},
		Log: logging.New(logging.Options{Service: "live", Writer: testWriter{t}, Level: slog.LevelError}),
	})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_ = ctl.Run(ctx)
	}()

	seed(ctx, t, probes, cluster.Consumer, "alpha", 2)
	waitPhase(ctx, t, probes, cluster.Consumer, "alpha", "Succeeded")
	seed(ctx, t, probes, cluster.Consumer, "beta", 1)
	waitPhase(ctx, t, probes, cluster.Consumer, "beta", "Succeeded")

	if ctl.Reconciles() == 0 {
		t.Fatal("the controller reconciled nothing")
	}
	t.Logf("reconciles %d, queue depth %d, cache age %s", ctl.Reconciles(), ctl.QueueDepth(), ctl.CacheAge().Round(time.Millisecond))
}

func seed(ctx context.Context, t *testing.T, probes *kcpstore.Resource[liveProbe], cluster, name string, steps int32) {
	t.Helper()
	obj := &liveProbe{APIVersion: livekcp.Group + "/" + livekcp.Version, Kind: livekcp.Kind}
	obj.Metadata.Name = name
	obj.Metadata.Namespace = "default"
	obj.Spec.Steps = steps
	if err := probes.Create(ctx, cluster, obj); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
}

func waitPhase(ctx context.Context, t *testing.T, probes *kcpstore.Resource[liveProbe], cluster, name, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		obj, err := probes.Get(ctx, ref.New(cluster, "default", name))
		if err == nil && obj.Status.Phase == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	obj, err := probes.Get(ctx, ref.New(cluster, "default", name))
	t.Fatalf("%s never reached %s: phase %q err %v", name, want, obj.Status.Phase, err)
}

type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(body []byte) (int, error) {
	w.t.Log(string(body))
	return len(body), nil
}
