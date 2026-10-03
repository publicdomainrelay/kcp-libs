package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/publicdomainrelay/kcp-libs/abc/queue"
	"github.com/publicdomainrelay/kcp-libs/abc/runref"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
	"github.com/publicdomainrelay/kcp-libs/factory/admission"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

const (
	namespace = "default"

	batchName = "nightly"

	itemCount = 5

	maxConcurrent = 2

	parentLabel = "example.computer/batch"

	roleLabel = "example.computer/role"
)

var widgets = schema.GroupVersionResource{Group: livekcp.Group, Version: livekcp.Version, Resource: livekcp.Resource}

type metadata struct {
	Name string `json:"name"`

	Namespace string `json:"namespace"`

	UID string `json:"uid"`

	ResourceVersion string `json:"resourceVersion"`

	Labels map[string]string `json:"labels,omitempty"`

	CreationTimestamp string `json:"creationTimestamp,omitempty"`
}

type widget struct {
	APIVersion string `json:"apiVersion,omitempty"`

	Kind string `json:"kind,omitempty"`

	Metadata metadata `json:"metadata"`

	Spec struct {
		ConcurrencyPolicy string `json:"concurrencyPolicy,omitempty"`

		MaxConcurrent *int32 `json:"maxConcurrent,omitempty"`

		Steps int32 `json:"steps,omitempty"`
	} `json:"spec"`

	Status struct {
		Phase string `json:"phase,omitempty"`

		Steps int32 `json:"steps,omitempty"`
	} `json:"status"`
}

func terminal(phase string) bool {
	return phase == "Succeeded" || phase == "Failed"
}

type source struct {
	items *kcpstore.Resource[widget]
}

func (s source) Parent(ctx context.Context, run queue.Run) (ref.Ref, bool, error) {
	obj, err := s.items.Get(ctx, run.Ref)
	if err != nil {
		return ref.Ref{}, false, err
	}
	parent := obj.Metadata.Labels[parentLabel]
	if parent == "" {
		return ref.Ref{}, false, nil
	}
	return ref.New(run.Ref.LogicalCluster, run.Ref.Namespace, parent), true, nil
}

func (s source) Runs(ctx context.Context, parent ref.Ref) ([]queue.Run, error) {
	all, err := s.items.List(ctx, parent.LogicalCluster)
	if err != nil {
		return nil, err
	}
	var out []queue.Run
	for i := range all {
		obj := &all[i]
		if obj.Metadata.Labels[parentLabel] != parent.Name {
			continue
		}
		created, _ := time.Parse(time.RFC3339Nano, obj.Metadata.CreationTimestamp)
		out = append(out, queue.Run{
			Ref:     ref.New(parent.LogicalCluster, obj.Metadata.Namespace, obj.Metadata.Name),
			Phase:   obj.Status.Phase,
			Created: created,
		})
	}
	return out, nil
}

func (s source) Capacity(ctx context.Context, parent ref.Ref) (queue.Capacity, *queue.Blocker, error) {
	obj, err := s.items.Get(ctx, parent)
	if err != nil {
		if kcpstore.IsNotFound(err) {
			return queue.Capacity{}, &queue.Blocker{Reason: "BatchMissing", Message: "the parent batch does not exist"}, nil
		}
		return queue.Capacity{}, nil, err
	}
	return queue.Capacity{Policy: queue.Policy(obj.Spec.ConcurrencyPolicy), MaxConcurrent: obj.Spec.MaxConcurrent}, nil, nil
}

type counters struct {
	starts int

	guardRefusals int

	woken int

	capacityWaits int

	peakRunning int

	waitMessage string
}

func (c *counters) wake(string, ref.Ref) {
	c.woken++
}

func Run(ctx context.Context, out io.Writer) error {
	cluster, err := livekcp.Start(ctx)
	if err != nil {
		return err
	}
	defer cluster.Stop()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.Server, RestConfig: cluster.Config})
	if err != nil {
		return err
	}
	resource := kcpstore.Of[widget](store, widgets)

	if err := seedBatch(ctx, resource, cluster.Consumer); err != nil {
		return err
	}
	for i := range itemCount {
		if err := seedItem(ctx, resource, cluster.Consumer, fmt.Sprintf("item-%d", i)); err != nil {
			return err
		}
	}

	tally := &counters{}
	admits := admission.New(admission.Options{
		Source:     source{items: resource},
		RunKind:    "item",
		ParentKind: "batch",
		Terminal:   terminal,
		Wake:       tally.wake,
	})
	started := runref.New(time.Minute)
	parent := ref.New(cluster.Consumer, namespace, batchName)
	deadline := time.Now().Add(2 * time.Minute)

	for pass := 0; time.Now().Before(deadline); pass++ {
		all, err := resource.List(ctx, cluster.Consumer)
		if err != nil {
			return err
		}
		running := 0
		pending := 0
		for i := range all {
			obj := &all[i]
			if obj.Metadata.Labels[roleLabel] != "item" || terminal(obj.Status.Phase) {
				continue
			}
			target := ref.New(cluster.Consumer, obj.Metadata.Namespace, obj.Metadata.Name)
			created, _ := time.Parse(time.RFC3339Nano, obj.Metadata.CreationTimestamp)
			decision, err := admits.Admit(ctx, queue.Run{Ref: target, Phase: obj.Status.Phase, Created: created})
			if err != nil {
				return err
			}
			if obj.Status.Phase == "Running" {
				running++
				if obj.Status.Steps < obj.Spec.Steps {
					if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, "Running", obj.Status.Steps+1); err != nil {
						return err
					}
					continue
				}
				if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, "Succeeded", obj.Status.Steps); err != nil {
					return err
				}
				if err := admits.Wake(ctx, parent); err != nil {
					return err
				}
				continue
			}
			pending++
			if decision.Gated && !decision.Allowed {
				tally.capacityWaits++
				if tally.waitMessage == "" {
					tally.waitMessage = fmt.Sprintf("%s waits: %s", obj.Metadata.Name, decision.Message)
				}
				continue
			}
			if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, "Running", 0); err != nil {
				return err
			}
			started.Record(target, "workload-"+obj.Metadata.Name, obj.Metadata.UID, time.Now())
			tally.starts++
			running++
			if known, found := started.Lookup(target); runref.AlreadyStarted(known, found, runref.Current{UID: obj.Metadata.UID}) {
				tally.guardRefusals++
			}
		}
		if running > tally.peakRunning {
			tally.peakRunning = running
		}
		if pending == 0 && running == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	final, err := resource.List(ctx, cluster.Consumer)
	if err != nil {
		return err
	}
	succeeded := 0
	for i := range final {
		if final[i].Metadata.Labels[roleLabel] == "item" && final[i].Status.Phase == "Succeeded" {
			succeeded++
		}
	}
	fmt.Fprintln(out, tally.waitMessage)
	fmt.Fprintf(out, "started %d of %d items, peak running %d\n", tally.starts, itemCount, tally.peakRunning)
	fmt.Fprintf(out, "waited at capacity on %d passes, woke %d queued runs\n", tally.capacityWaits, tally.woken)
	fmt.Fprintf(out, "the duplicate-start guard refused %d stale starts, leases held %d, succeeded %d\n",
		tally.guardRefusals, admits.Leases().Len(), succeeded)
	return nil
}

func seedBatch(ctx context.Context, resource *kcpstore.Resource[widget], cluster string) error {
	_ = resource.Delete(ctx, ref.New(cluster, namespace, batchName))
	obj := &widget{APIVersion: livekcp.APIVersion, Kind: livekcp.Kind}
	obj.Metadata.Name = batchName
	obj.Metadata.Namespace = namespace
	obj.Metadata.Labels = map[string]string{roleLabel: "batch"}
	obj.Spec.ConcurrencyPolicy = "Forbid"
	obj.Spec.MaxConcurrent = ptr(int32(maxConcurrent))
	return resource.Create(ctx, cluster, obj)
}

func seedItem(ctx context.Context, resource *kcpstore.Resource[widget], cluster, name string) error {
	_ = resource.Delete(ctx, ref.New(cluster, namespace, name))
	obj := &widget{APIVersion: livekcp.APIVersion, Kind: livekcp.Kind}
	obj.Metadata.Name = name
	obj.Metadata.Namespace = namespace
	obj.Metadata.Labels = map[string]string{parentLabel: batchName, roleLabel: "item"}
	obj.Spec.Steps = 1
	return resource.Create(ctx, cluster, obj)
}

func writeStatus(ctx context.Context, resource *kcpstore.Resource[widget], target ref.Ref, version, phase string, steps int32) error {
	patch, err := statuspatch.Merge(map[string]any{"phase": phase, "steps": steps})
	if err != nil {
		return err
	}
	return resource.PatchStatus(ctx, target.WithResourceVersion(version), patch)
}

func ptr[T any](value T) *T {
	return &value
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-admission:", err)
		os.Exit(1)
	}
}
