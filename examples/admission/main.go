package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/publicdomainrelay/kcp-libs/abc/queue"
	"github.com/publicdomainrelay/kcp-libs/abc/runref"
	"github.com/publicdomainrelay/kcp-libs/common/deno"
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

type itemSpec struct {
	Steps int32 `json:"steps,omitempty"`
}

type itemStatus struct {
	Phase string `json:"phase,omitempty"`

	Steps int32 `json:"steps,omitempty"`
}

type batchSpec struct {
	ConcurrencyPolicy string `json:"concurrencyPolicy,omitempty"`

	MaxConcurrent *int32 `json:"maxConcurrent,omitempty"`
}

type batchStatus struct{}

type batch = livekcp.Object[batchSpec, batchStatus]

type item = livekcp.Object[itemSpec, itemStatus]

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

	woken int

	capacityWaits int

	peakRunning int

	waitMessage string

	refused int

	allowed int
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
	resource := kcpstore.Of[item](store, livekcp.WidgetGVR)

	batches := kcpstore.Of[batch](store, livekcp.WidgetGVR)
	batchObject := livekcp.NewObject[batchSpec, batchStatus](namespace, batchName)
	batchObject.Metadata.Labels = map[string]string{roleLabel: "batch"}
	batchObject.Spec.ConcurrencyPolicy = string(deno.ConcurrencyAllow)
	maxItems := int32(maxConcurrent)
	batchObject.Spec.MaxConcurrent = &maxItems
	if err := livekcp.Seed(ctx, cluster.ConsumerCluster, batches, namespace, batchName, batchObject); err != nil {
		return err
	}
	for i := range itemCount {
		item := livekcp.NewObject[itemSpec, itemStatus](namespace, fmt.Sprintf("item-%d", i))
		item.Metadata.Labels = map[string]string{parentLabel: batchName, roleLabel: "item"}
		item.Spec.Steps = 1
		if err := livekcp.Seed(ctx, cluster.ConsumerCluster, resource, namespace, item.Metadata.Name, item); err != nil {
			return err
		}
	}

	tally := &counters{}
	admits := admission.New(admission.Options{
		Source:     source{items: resource, batches: batches},
		RunKind:    "item",
		ParentKind: "batch",
		Terminal:   deno.TerminalPolicyWorkflow,
		Wake:       tally.wake,
	})
	started := runref.New(time.Minute)
	parent := ref.New(cluster.ConsumerCluster, namespace, batchName)
	deadline := time.Now().Add(2 * time.Minute)
	startedAt := time.Now()
	passes := 0

	for pass := 0; time.Now().Before(deadline); pass++ {
		passes = pass
		all, err := resource.List(ctx, cluster.ConsumerCluster)
		if err != nil {
			return err
		}
		observedRunning := 0
		remaining := 0
		for i := range all {
			obj := &all[i]
			if obj.Metadata.Labels[roleLabel] != "item" || deno.TerminalPolicyWorkflow(obj.Status.Phase) {
				continue
			}
			remaining++
			if obj.Status.Phase == string(deno.PhaseRunning) {
				observedRunning++
			}
			target := ref.New(cluster.ConsumerCluster, obj.Metadata.Namespace, obj.Metadata.Name)
			created, _ := time.Parse(time.RFC3339Nano, obj.Metadata.CreationTimestamp)
			decision, err := admits.Admit(ctx, queue.Run{Ref: target, Phase: obj.Status.Phase, Created: created})
			if err != nil {
				return err
			}
			if obj.Status.Phase == string(deno.PhaseRunning) {
				if obj.Status.Steps < obj.Spec.Steps {
					if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, string(deno.PhaseRunning), obj.Status.Steps+1); err != nil {
						return err
					}
					continue
				}
				if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, string(deno.PhaseSucceeded), obj.Status.Steps); err != nil {
					return err
				}
				if err := admits.Wake(ctx, parent); err != nil {
					return err
				}
				continue
			}
			if decision.Gated && !decision.Allowed {
				tally.capacityWaits++
				if tally.waitMessage == "" {
					tally.waitMessage = fmt.Sprintf("%s waits: %s", obj.Metadata.Name, decision.Message)
				}
				continue
			}
			if err := writeStatus(ctx, resource, target, obj.Metadata.ResourceVersion, string(deno.PhaseRunning), 0); err != nil {
				return err
			}
			started.Record(target, "workload-"+obj.Metadata.Name, obj.Metadata.UID, time.Now())
			tally.starts++
			stale := runref.Current{UID: obj.Metadata.UID}
			if known, found := started.Lookup(target); runref.AlreadyStarted(known, found, stale) {
				tally.refused++
			}
			if known, found := started.Lookup(target); !runref.AlreadyStarted(known, found, runref.Current{UID: obj.Metadata.UID, Retries: 1}) {
				tally.allowed++
			}
			if known, found := started.Lookup(target); !runref.AlreadyStarted(known, found, runref.Current{UID: "recreated"}) {
				tally.allowed++
			}
		}
		if observedRunning > tally.peakRunning {
			tally.peakRunning = observedRunning
		}
		if remaining == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	final, err := resource.List(ctx, cluster.ConsumerCluster)
	if err != nil {
		return err
	}
	succeeded := 0
	for i := range final {
		if final[i].Metadata.Labels[roleLabel] == "item" && final[i].Status.Phase == string(deno.PhaseSucceeded) {
			succeeded++
		}
	}
	limit, unlimited := queue.Limit(queue.Policy(deno.ConcurrencyAllow), &maxItems)
	fmt.Fprintln(out, tally.waitMessage)
	fmt.Fprintf(out, "the batch allows %d at once and %d items were created\n", limit, itemCount)
	fmt.Fprintf(out, "started %d of %d items, peak observed running %d, waited at capacity on %d passes\n",
		tally.starts, itemCount, tally.peakRunning, tally.capacityWaits)
	fmt.Fprintf(out, "drained in %d passes over %s\n", passes, time.Since(startedAt).Round(time.Millisecond))
	fmt.Fprintf(out, "woke %d queued runs, leases held %d, succeeded %d, unlimited %v\n",
		tally.woken, admits.Leases().Len(), succeeded, unlimited)
	fmt.Fprintf(out, "the duplicate-start guard refused %d stale copies and allowed %d legitimate starts\n",
		tally.refused, tally.allowed)
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
