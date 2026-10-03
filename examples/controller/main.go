package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/abc/driver"
	"github.com/publicdomainrelay/kcp-libs/abc/reconcile"
	abcstore "github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/condition"
	"github.com/publicdomainrelay/kcp-libs/common/deno"
	"github.com/publicdomainrelay/kcp-libs/common/env"
	"github.com/publicdomainrelay/kcp-libs/common/logging"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
	"github.com/publicdomainrelay/kcp-libs/factory/controller"
	"github.com/publicdomainrelay/kcp-libs/fakekcp"
	"github.com/publicdomainrelay/kcp-libs/impl/exportwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/impl/metrics"
)

const (
	apiVersion = "example.computer/v1alpha1"

	kind = "Widget"

	providerWorkspace = "root:deno-provider"

	export = "denoruntime"

	workspaceID = "2j35eh7jjhsc8ny9"

	workspacePath = "root:alice"

	resource = "widgets"

	parentLabel = "example.computer/widget-group"
)

var widgets = schema.GroupVersionResource{Group: "example.computer", Version: "v1alpha1", Resource: resource}

type metadata struct {
	Name string `json:"name"`

	Namespace string `json:"namespace"`

	UID string `json:"uid"`

	ResourceVersion string `json:"resourceVersion"`

	Labels map[string]string `json:"labels"`

	Finalizers []string `json:"finalizers"`
}

type spec struct {
	Steps int32 `json:"steps"`

	Parent string `json:"parent"`
}

type status struct {
	Phase string `json:"phase,omitempty"`

	Observed int32 `json:"observed,omitempty"`

	Siblings int32 `json:"siblings,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type widget struct {
	APIVersion string `json:"apiVersion"`

	Kind string `json:"kind"`

	Metadata metadata `json:"metadata"`

	Spec spec `json:"spec"`

	Status status `json:"status"`
}

type observed struct {
	Widget widget

	Siblings int32

	Now time.Time
}

func decide(_ context.Context, o observed) (reconcile.Result[status], error) {
	result := reconcile.Result[status]{Phase: o.Widget.Status.Phase, Status: o.Widget.Status}
	result.Status.Siblings = o.Siblings
	switch o.Widget.Status.Phase {
	case "":
		result.Phase = string(deno.PhasePending)
		result.Add(reconcile.OpStart)
	case string(deno.PhasePending):
		result.Phase = string(deno.PhaseRunning)
		result.Status.Observed = 0
	case string(deno.PhaseRunning):
		result.Status.Observed++
		if result.Status.Observed >= o.Widget.Spec.Steps {
			result.Phase = string(deno.PhaseSucceeded)
		}
	default:
		return result, nil
	}
	result.Status.Conditions = condition.Copy(result.Status.Conditions)
	if result.Phase == string(deno.PhaseSucceeded) {
		condition.SetTrue(&result.Status.Conditions, 1, deno.ConditionComplete, "Complete", "every step ran")
	} else {
		condition.SetFalse(&result.Status.Conditions, 1, deno.ConditionComplete, "Running", "steps remain")
	}
	result.RequeueAfter = time.Millisecond
	return result, nil
}

func handlerFor(set *cache.Set, resource *kcpstore.Resource[widget]) driver.Handler {
	return driver.HandlerFunc(func(ctx context.Context, key driver.Key) (time.Duration, bool, error) {
		obj, err := resource.Get(ctx, key.Ref)
		if err != nil {
			if kcpstore.IsNotFound(err) {
				return 0, true, nil
			}
			return 0, false, err
		}
		group := obj.Metadata.Labels[parentLabel]
		siblings := int32(len(set.ByIndex("widget", cache.ByClusterParent, ref.Key(key.Ref.LogicalCluster, key.Ref.Namespace, group))))

		result, err := reconcile.Decider(decide).Reconcile(ctx, observed{Widget: *obj, Siblings: siblings, Now: time.Now()})
		if err != nil {
			return 0, false, err
		}
		terminal := deno.TerminalPolicyWorkflow(result.Phase)
		next := status{
			Phase:      result.Phase,
			Observed:   result.Status.Observed,
			Siblings:   result.Status.Siblings,
			Conditions: result.Status.Conditions,
		}
		if !abcstore.Same(next, obj.Status) {
			patch, err := statuspatch.Merge(map[string]any{
				"phase":      next.Phase,
				"observed":   next.Observed,
				"siblings":   next.Siblings,
				"conditions": statuspatch.Optional(next.Conditions),
			})
			if err != nil {
				return 0, false, err
			}
			target := key.Ref.WithResourceVersion(obj.Metadata.ResourceVersion)
			if err := resource.PatchStatus(ctx, target, patch); err != nil {
				return 0, false, err
			}
		}
		return result.RequeueAfter, terminal, nil
	})
}

func Run(ctx context.Context, out io.Writer) error {
	logger := logging.New(logging.Options{Service: "example-controller", Writer: out, Level: slog.LevelError})
	cluster, err := fakekcp.New()
	if err != nil {
		return err
	}
	defer cluster.Close()

	cluster.SetClusterPath(workspaceID, workspacePath)
	cluster.AddEndpointSlice(export, cluster.URL()+"/services/apiexport/"+providerWorkspace+"/"+export)

	ctx, cancel := context.WithTimeout(ctx, env.OrDuration("EXAMPLE_TIMEOUT", 30*time.Second))
	defer cancel()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.URL(), RestConfig: &rest.Config{Host: cluster.URL()}})
	if err != nil {
		return err
	}
	endpoints, err := exportwatch.Await(ctx, exportwatch.Options{
		Config:            store.Config(),
		Host:              cluster.URL(),
		ProviderWorkspace: providerWorkspace,
		Exports:           []string{export},
		Log:               logger,
	})
	if err != nil {
		return err
	}
	bases := exportwatch.Paths(endpoints, export)
	if len(bases) == 0 {
		return fmt.Errorf("example: no virtual workspace URL was published for %s", export)
	}
	fmt.Fprintf(out, "discovered %s at %s\n", export, bases[0])

	seed(cluster, "alpha", 1)
	seed(cluster, "beta", 1)

	resource := kcpstore.Of[widget](store, widgets)
	var _ abcstore.Resource[widget] = resource
	registry := metrics.New("example")
	watched := cache.NewSet()
	handler := handlerFor(watched, resource)

	ctl, err := controller.New(controller.Options{
		Config:    store.Config(),
		Bases:     bases,
		Resources: []informerwatch.Resource{{Kind: "widget", GVR: widgets}},
		Indexers:  cache.IndexersFor(parentLabel, "", ""),
		Set:       watched,
		Handler:   handler,
		Policy: driver.Policy{
			Interval:          5 * time.Millisecond,
			MinTransitionPoll: time.Millisecond,
			ClampKinds:        map[string]bool{"widget": true},
		},
		Metrics: registry,
		Log:     logger,
	})
	if err != nil {
		return err
	}
	go func() {
		_ = ctl.Run(ctx)
	}()

	for _, name := range []string{"alpha", "beta"} {
		if err := cluster.WaitFor(ctx, func() bool { return succeeded(cluster, name) }); err != nil {
			return err
		}
		fmt.Fprintf(out, "%s reached %s after %d passes\n", name, phaseOf(cluster, name), observedOf(cluster, name))
	}

	seed(cluster, "gamma", 1)
	if err := cluster.WaitFor(ctx, func() bool { return succeeded(cluster, "gamma") }); err != nil {
		return err
	}
	fmt.Fprintf(out, "gamma reached %s from the watch, seeing %d siblings in the cache\n",
		phaseOf(cluster, "gamma"), siblingsOf(cluster, "gamma"))

	fmt.Fprintf(out, "queue depth %d, cache age %s, patches %d\n", ctl.QueueDepth(), ctl.CacheAge().Round(time.Millisecond), cluster.Patches())
	registry.Render(out)
	return nil
}

func seed(cluster *fakekcp.Cluster, name string, steps int32) {
	obj := fakekcp.Object(apiVersion, kind, "default", name)
	fakekcp.WithLabel(obj, parentLabel, "group-a")
	fakekcp.WithSpec(obj, map[string]any{"steps": steps, "parent": "group-a"})
	cluster.Create(workspaceID, "default", resource, obj)
}

func bodyOf(cluster *fakekcp.Cluster, name string) map[string]any {
	body, found := cluster.Get(workspaceID, "default", resource, name)
	if !found {
		return nil
	}
	return body
}

func phaseOf(cluster *fakekcp.Cluster, name string) string {
	return fakekcp.PhaseOf(bodyOf(cluster, name))
}

func succeeded(cluster *fakekcp.Cluster, name string) bool {
	return deno.TerminalPolicyWorkflow(phaseOf(cluster, name))
}

func observedOf(cluster *fakekcp.Cluster, name string) int {
	value, _ := fakekcp.StatusOf(bodyOf(cluster, name))["observed"].(float64)
	return int(value)
}

func siblingsOf(cluster *fakekcp.Cluster, name string) int {
	value, _ := fakekcp.StatusOf(bodyOf(cluster, name))["siblings"].(float64)
	return int(value)
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-controller:", err)
		os.Exit(1)
	}
}
