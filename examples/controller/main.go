package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/abc/reconcile"
	abcstore "github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/condition"
	"github.com/publicdomainrelay/kcp-libs/common/denocomputer"
	"github.com/publicdomainrelay/kcp-libs/common/logging"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/factory/controller"
	"github.com/publicdomainrelay/kcp-libs/impl/exportwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/informerwatch"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/impl/metrics"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "default"

	parentLabel = "example.computer/group"

	groupName = "group-a"
)

type spec struct {
	Steps int32 `json:"steps"`
}

type status struct {
	Phase string `json:"phase,omitempty"`

	Observed int32 `json:"observed,omitempty"`

	Siblings int32 `json:"siblings,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type widget = livekcp.Object[spec, status]

type observed struct {
	Widget widget

	Siblings int32
}

func decide(_ context.Context, o observed) (reconcile.Result[status], error) {
	result := reconcile.Result[status]{Phase: o.Widget.Status.Phase, Status: o.Widget.Status}
	result.Status.Siblings = o.Siblings
	switch o.Widget.Status.Phase {
	case "":
		result.Phase = string(denocomputer.PhasePending)
		result.Add(reconcile.KindStart)
	case string(denocomputer.PhasePending):
		result.Phase = string(denocomputer.PhaseRunning)
		result.Status.Observed = 0
	case string(denocomputer.PhaseRunning):
		result.Status.Observed++
		if result.Status.Observed >= o.Widget.Spec.Steps {
			result.Phase = string(denocomputer.PhaseSucceeded)
		}
	default:
		return result, nil
	}
	result.Status.Conditions = condition.Copy(result.Status.Conditions)
	if result.Phase == string(denocomputer.PhaseSucceeded) {
		condition.SetTrue(&result.Status.Conditions, 1, denocomputer.ConditionComplete, "Complete", "every step ran")
	} else {
		condition.SetFalse(&result.Status.Conditions, 1, denocomputer.ConditionComplete, "Running", "steps remain")
	}
	result.RequeueAfter = time.Millisecond
	return result, nil
}

func handlerFor(set *cache.Set, resource *kcpstore.Resource[widget]) reconcile.Handler {
	return reconcile.Bridge[observed, status]{
		Read: func(ctx context.Context, key reconcile.Key) (observed, error) {
			obj, err := resource.Get(ctx, key.Ref)
			if err != nil {
				if kcpstore.IsNotFound(err) {
					return observed{}, reconcile.ErrGone
				}
				return observed{}, err
			}
			group := obj.Metadata.Labels[parentLabel]
			siblings := int32(len(set.ByIndex("widget", cache.ByClusterParent,
				ref.Key(key.Ref.LogicalCluster, key.Ref.Namespace, group))))
			return observed{Widget: *obj, Siblings: siblings}, nil
		},
		Decider: reconcile.Func[observed, status](decide),
		Apply: func(ctx context.Context, key reconcile.Key, o observed, result reconcile.Result[status]) error {
			next := status{
				Phase:      result.Phase,
				Observed:   result.Status.Observed,
				Siblings:   result.Status.Siblings,
				Conditions: result.Status.Conditions,
			}
			if abcstore.Unchanged(next, o.Widget.Status) {
				return nil
			}
			patch, err := reconcile.Patch(reconcile.Result[status]{Status: next})
			if err != nil {
				return err
			}
			return resource.PatchStatus(ctx, key.Ref.WithResourceVersion(o.Widget.Metadata.ResourceVersion), patch)
		},
		Terminal: denocomputer.TerminalPolicyWorkflow,
	}
}

func Run(ctx context.Context, out io.Writer) error {
	logger := logging.New(logging.Options{Service: "example-controller", Writer: out, Level: slog.LevelError})
	cluster, err := livekcp.Start(ctx)
	if err != nil {
		return err
	}
	defer cluster.Stop()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.Server, RestConfig: cluster.Config})
	if err != nil {
		return err
	}
	endpoints, err := exportwatch.Await(ctx, exportwatch.Options{
		Config:            store.Config(),
		Host:              cluster.Server,
		ProviderWorkspace: cluster.ProviderCluster,
		Exports:           []string{livekcp.Export},
		Log:               logger,
	})
	if err != nil {
		return err
	}
	bases := exportwatch.Paths(endpoints, livekcp.Export)
	if len(bases) == 0 {
		return fmt.Errorf("example: kcp published no virtual workspace URL for %s", livekcp.Export)
	}
	fmt.Fprintf(out, "kcp published %s at %s\n", livekcp.Export, bases[0])

	resource := kcpstore.Of[widget](store, livekcp.WidgetGVR)
	var _ abcstore.Resource[widget] = resource
	registry := metrics.New("example")
	watched := cache.NewSet()
	handler := handlerFor(watched, resource)

	for _, name := range []string{"alpha", "beta"} {
		if err := seed(ctx, cluster.ConsumerCluster, resource, name, 1); err != nil {
			return err
		}
	}

	ctl, err := controller.New(controller.Options{
		Config:   store.Config(),
		Sources:  []informerwatch.Source{{Base: bases[0], Resources: []informerwatch.Resource{{Kind: "widget", GVR: livekcp.WidgetGVR}}}},
		Indexers: cache.IndexersFor(parentLabel, "", ""),
		Set:      watched,
		Handler:  handler,
		Policy: reconcile.Policy{
			Interval:          time.Second,
			MinTransitionPoll: 10 * time.Millisecond,
			ClampKinds:        map[string]bool{"widget": true},
		},
		Metrics: registry,
		Log:     logger,
	})
	if err != nil {
		return err
	}
	watchCtx, stop := context.WithCancel(ctx)
	defer stop()
	go func() {
		_ = ctl.Run(watchCtx)
	}()

	for _, name := range []string{"alpha", "beta"} {
		obj, err := livekcp.WaitFor(ctx, cluster.ConsumerCluster, resource, namespace, name, succeeded)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s reached %s after %d passes\n", name, obj.Status.Phase, obj.Status.Observed)
	}

	if err := seed(ctx, cluster.ConsumerCluster, resource, "gamma", 1); err != nil {
		return err
	}
	gamma, err := livekcp.WaitFor(ctx, cluster.ConsumerCluster, resource, namespace, "gamma", succeeded)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "gamma reached %s from the watch, having seen %d of the group in the cache\n",
		gamma.Status.Phase, gamma.Status.Siblings)
	fmt.Fprintf(out, "reconciles %d, queue depth %d, cache age %s\n",
		ctl.Reconciles(), ctl.QueueDepth(), ctl.CacheAge().Round(time.Millisecond))
	registry.Render(out)
	return nil
}

func succeeded(obj widget) bool {
	return obj.Status.Phase == string(denocomputer.PhaseSucceeded)
}

func seed(ctx context.Context, cluster string, resource *kcpstore.Resource[widget], name string, steps int32) error {
	obj := livekcp.NewObject[spec, status](namespace, name)
	obj.Metadata.Labels = map[string]string{parentLabel: groupName}
	obj.Spec.Steps = steps
	return livekcp.Seed(ctx, cluster, resource, namespace, name, obj)
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-controller:", err)
		os.Exit(1)
	}
}
