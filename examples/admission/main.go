package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/abc/queue"
	"github.com/publicdomainrelay/kcp-libs/abc/runref"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
	"github.com/publicdomainrelay/kcp-libs/factory/admission"
	"github.com/publicdomainrelay/kcp-libs/fakekcp"
	"github.com/publicdomainrelay/kcp-libs/impl/kcpstore"
)

const (
	apiVersion = "example.computer/v1alpha1"

	batchKind = "Batch"

	itemKind = "Item"

	batchResourceName = "batches"

	itemResourceName = "items"

	workspaceID = "2j35eh7jjhsc8ny9"

	batchName = "nightly"

	itemCount = 5

	parentLabel = "example.computer/batch"
)

var (
	batchesResource = schema.GroupVersionResource{Group: "example.computer", Version: "v1alpha1", Resource: batchResourceName}

	itemsResource = schema.GroupVersionResource{Group: "example.computer", Version: "v1alpha1", Resource: itemResourceName}
)

type objectMeta struct {
	Name string `json:"name"`

	Namespace string `json:"namespace"`

	UID string `json:"uid"`

	ResourceVersion string `json:"resourceVersion"`

	Labels map[string]string `json:"labels"`

	CreationTimestamp string `json:"creationTimestamp"`
}

type batch struct {
	Metadata objectMeta `json:"metadata"`

	Spec struct {
		ConcurrencyPolicy string `json:"concurrencyPolicy"`

		MaxConcurrent *int32 `json:"maxConcurrent"`
	} `json:"spec"`
}

type item struct {
	Metadata objectMeta `json:"metadata"`

	Spec struct {
		Steps int32 `json:"steps"`
	} `json:"spec"`

	Status struct {
		Phase string `json:"phase"`

		Steps int32 `json:"steps"`
	} `json:"status"`
}

func terminal(phase string) bool {
	return phase == "Succeeded" || phase == "Failed"
}

type source struct {
	items *kcpstore.Resource[item]

	batches *kcpstore.Resource[batch]
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
	obj, err := s.batches.Get(ctx, parent)
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

	atCapacity int

	peakRunning int
}

func (c *counters) wake(string, ref.Ref) {
	c.woken++
}

func Run(ctx context.Context, out io.Writer) error {
	cluster, err := fakekcp.New()
	if err != nil {
		return err
	}
	defer cluster.Close()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	store, err := kcpstore.New(kcpstore.Options{Host: cluster.URL(), RestConfig: &rest.Config{Host: cluster.URL()}})
	if err != nil {
		return err
	}
	items := kcpstore.Of[item](store, itemsResource)
	batches := kcpstore.Of[batch](store, batchesResource)

	batchBody := fakekcp.Object(apiVersion, batchKind, "default", batchName)
	fakekcp.WithSpec(batchBody, map[string]any{"concurrencyPolicy": "Forbid", "maxConcurrent": 2})
	cluster.Create(workspaceID, "default", batchResourceName, batchBody)

	base := time.Now().Add(-time.Hour)
	for i := range itemCount {
		body := fakekcp.Object(apiVersion, itemKind, "default", fmt.Sprintf("item-%d", i))
		fakekcp.WithLabel(body, parentLabel, batchName)
		fakekcp.WithSpec(body, map[string]any{"steps": 1})
		fakekcp.WithCreated(body, base.Add(time.Duration(i)*time.Second))
		cluster.Create(workspaceID, "default", itemResourceName, body)
	}

	tally := &counters{}
	admits := admission.New(admission.Options{
		Source:     source{items: items, batches: batches},
		RunKind:    "item",
		ParentKind: "batch",
		Terminal:   terminal,
		Wake:       tally.wake,
	})
	started := runref.New(time.Minute)
	parent := ref.New(workspaceID, "default", batchName)

	for pass := range 500 {
		all, err := items.List(ctx, workspaceID)
		if err != nil {
			return err
		}
		running := 0
		pending := 0
		for i := range all {
			obj := &all[i]
			if terminal(obj.Status.Phase) {
				continue
			}
			target := ref.New(workspaceID, obj.Metadata.Namespace, obj.Metadata.Name)
			created, _ := time.Parse(time.RFC3339Nano, obj.Metadata.CreationTimestamp)
			decision, err := admits.Admit(ctx, queue.Run{Ref: target, Phase: obj.Status.Phase, Created: created})
			if err != nil {
				return err
			}
			if obj.Status.Phase == "Running" {
				running++
				if obj.Status.Steps < obj.Spec.Steps {
					if err := writeStatus(ctx, items, target, obj.Metadata.ResourceVersion, "Running", obj.Status.Steps+1); err != nil {
						return err
					}
					continue
				}
				if err := writeStatus(ctx, items, target, obj.Metadata.ResourceVersion, "Succeeded", obj.Status.Steps); err != nil {
					return err
				}
				if err := admits.Wake(ctx, parent); err != nil {
					return err
				}
				continue
			}
			pending++
			if decision.Gated && !decision.Allowed {
				tally.atCapacity++
				if pass == 0 && obj.Metadata.Name == "item-2" {
					fmt.Fprintf(out, "%s waits: %s\n", obj.Metadata.Name, decision.Message)
				}
				continue
			}
			if err := writeStatus(ctx, items, target, obj.Metadata.ResourceVersion, "Running", 0); err != nil {
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
		time.Sleep(time.Millisecond)
	}

	final, err := items.List(ctx, workspaceID)
	if err != nil {
		return err
	}
	succeeded := 0
	for i := range final {
		if final[i].Status.Phase == "Succeeded" {
			succeeded++
		}
	}
	fmt.Fprintf(out, "started %d of %d items, peak running %d\n", tally.starts, itemCount, tally.peakRunning)
	fmt.Fprintf(out, "waited at capacity on %d passes, woke %d queued runs\n", tally.atCapacity, tally.woken)
	fmt.Fprintf(out, "the duplicate-start guard refused %d stale starts, leases held %d, succeeded %d\n",
		tally.guardRefusals, admits.Leases().Len(), succeeded)
	return nil
}

func writeStatus(ctx context.Context, resource *kcpstore.Resource[item], target ref.Ref, version, phase string, steps int32) error {
	patch, err := statuspatch.Merge(map[string]any{"phase": phase, "steps": steps})
	if err != nil {
		return err
	}
	return resource.PatchStatus(ctx, target.WithResourceVersion(version), patch)
}

func main() {
	if err := Run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "example-admission:", err)
		os.Exit(1)
	}
}
