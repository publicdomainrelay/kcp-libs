package queue

import (
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/deno"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

func int32Ptr(v int32) *int32 { return &v }

func run(name string, phase string, seconds int) Run {
	return Run{
		Ref:     ref.New("root:alice", "default", name),
		Phase:   phase,
		Created: time.Unix(int64(seconds), 0).UTC(),
	}
}

func TestLimit(t *testing.T) {
	cases := []struct {
		policy    Policy
		max       *int32
		limit     int32
		unlimited bool
	}{
		{PolicyAllow, nil, 0, true},
		{PolicyAllow, int32Ptr(0), 0, true},
		{PolicyAllow, int32Ptr(-1), 0, true},
		{PolicyAllow, int32Ptr(3), 3, false},
		{PolicyForbid, nil, 1, false},
		{PolicyForbid, int32Ptr(7), 1, false},
		{PolicyReplace, int32Ptr(7), 1, false},
		{Policy(""), int32Ptr(7), 1, false},
	}
	for _, tc := range cases {
		limit, unlimited := Limit(tc.policy, tc.max)
		if limit != tc.limit || unlimited != tc.unlimited {
			t.Fatalf("Limit(%q, %v) = (%d, %v), want (%d, %v)", tc.policy, tc.max, limit, unlimited, tc.limit, tc.unlimited)
		}
	}
}

func TestDecision(t *testing.T) {
	if allowed, _, _ := Decision(PolicyAllow, nil, 100, 5); !allowed {
		t.Fatal("allow with no maximum must admit everything")
	}
	if allowed, _, _ := Decision(PolicyAllow, int32Ptr(2), 1, 0); !allowed {
		t.Fatal("allow with a free slot must admit")
	}
	allowed, reason, message := Decision(PolicyAllow, int32Ptr(2), 2, 0)
	if allowed || reason != deno.ReasonAtCapacity {
		t.Fatalf("allow at capacity = (%v, %q), want (false, %q)", allowed, reason, deno.ReasonAtCapacity)
	}
	if message == "" {
		t.Fatal("a refusal must carry a message")
	}
	if allowed, _, _ := Decision(PolicyForbid, nil, 0, 0); !allowed {
		t.Fatal("forbid with one free slot must admit the first")
	}
	if allowed, _, _ := Decision(PolicyForbid, nil, 0, 1); allowed {
		t.Fatal("forbid must refuse the second")
	}
	if allowed, _, _ := Decision(PolicyForbid, nil, 1, 0); allowed {
		t.Fatal("forbid must refuse while one is running")
	}
}

func TestPlanForbidOrdersOldestFirst(t *testing.T) {
	runs := []Run{
		run("c", string(deno.PhasePending), 3),
		run("a", string(deno.PhasePending), 1),
		run("b", string(deno.PhasePending), 2),
	}
	planned := PlanIndex(runs, Capacity{Policy: PolicyForbid}, nil, 0)
	if !planned[ref.New("root:alice", "default", "a")].Allowed {
		t.Fatal("the oldest pending run must be admitted")
	}
	if planned[ref.New("root:alice", "default", "b")].Allowed {
		t.Fatal("the second pending run must wait")
	}
	if planned[ref.New("root:alice", "default", "b")].Position != 1 {
		t.Fatal("positions must follow oldest-first order")
	}
	if planned[ref.New("root:alice", "default", "c")].Message == "" {
		t.Fatal("a waiting run must carry the at-capacity message")
	}
}

func TestPlanCountsRunningAndTerminal(t *testing.T) {
	runs := []Run{
		run("running", string(deno.PhaseRunning), 1),
		run("done", string(deno.PhaseSucceeded), 2),
		run("pending", string(deno.PhasePending), 3),
	}
	planned := PlanIndex(runs, Capacity{Policy: PolicyAllow, MaxConcurrent: int32Ptr(1)}, nil, 0)
	if planned[ref.New("root:alice", "default", "running")].Gated {
		t.Fatal("a running run is not gated")
	}
	if planned[ref.New("root:alice", "default", "done")].Gated {
		t.Fatal("a terminal run is not gated")
	}
	waiting := planned[ref.New("root:alice", "default", "pending")]
	if waiting.Allowed {
		t.Fatal("a pending run must wait behind the running one")
	}
	if waiting.Active != 1 {
		t.Fatalf("active = %d, want 1", waiting.Active)
	}
}

func TestPlanReservedCountsTowardActive(t *testing.T) {
	runs := []Run{run("pending", string(deno.PhasePending), 1)}
	planned := PlanIndex(runs, Capacity{Policy: PolicyAllow, MaxConcurrent: int32Ptr(1)}, nil, 1)
	if planned[ref.New("root:alice", "default", "pending")].Allowed {
		t.Fatal("a reserved slot must not be handed out twice")
	}
}

func TestPlanReplaceAdmitsNewestAndPreempts(t *testing.T) {
	runs := []Run{
		run("old", string(deno.PhasePending), 1),
		run("middle", string(deno.PhaseRunning), 2),
		run("new", string(deno.PhasePending), 3),
	}
	planned := PlanIndex(runs, Capacity{Policy: PolicyReplace}, nil, 0)
	newest := planned[ref.New("root:alice", "default", "new")]
	if !newest.Allowed {
		t.Fatal("replace must admit the newest pending run")
	}
	if len(newest.Preempt) != 2 {
		t.Fatalf("preempt = %d refs, want 2", len(newest.Preempt))
	}
	if planned[ref.New("root:alice", "default", "old")].Reason != deno.ReasonSuperseded {
		t.Fatal("an older pending run is superseded")
	}
	if planned[ref.New("root:alice", "default", "middle")].Gated {
		t.Fatal("a running run is not gated, and is still preempted")
	}
}

func TestPlanBlockerGatesEveryPendingRun(t *testing.T) {
	runs := []Run{
		run("a", string(deno.PhasePending), 1),
		run("b", string(deno.PhasePending), 2),
	}
	blocker := &Blocker{Reason: deno.ReasonEngineNotReady, Message: "no endpoint"}
	planned := PlanIndex(runs, Capacity{Policy: PolicyForbid}, blocker, 0)
	for _, candidate := range runs {
		admission := planned[candidate.Ref]
		if !admission.Gated || admission.Allowed {
			t.Fatalf("%s: gated=%v allowed=%v, want gated and refused", candidate.Ref.Name, admission.Gated, admission.Allowed)
		}
		if admission.Reason != deno.ReasonEngineNotReady {
			t.Fatalf("%s: reason = %q", candidate.Ref.Name, admission.Reason)
		}
	}
}

func TestWakeListTakesCapacityManyOldestFirst(t *testing.T) {
	runs := []Run{
		run("running", string(deno.PhaseRunning), 1),
		run("done", string(deno.PhaseSucceeded), 2),
		run("c", string(deno.PhasePending), 5),
		run("a", string(deno.PhasePending), 3),
		run("b", string(deno.PhasePending), 4),
	}
	woken := WakeList(runs, deno.TerminalPolicyWorkflow, 2, false)
	if len(woken) != 2 {
		t.Fatalf("woken = %d, want 2", len(woken))
	}
	if woken[0].Name != "a" || woken[1].Name != "b" {
		t.Fatalf("woken = %v, want the two oldest pending runs", []string{woken[0].Name, woken[1].Name})
	}
	if all := WakeList(runs, deno.TerminalPolicyWorkflow, 0, true); len(all) != 3 {
		t.Fatalf("unlimited woken = %d, want 3", len(all))
	}
}

func TestGroupByParentSkipsUnparented(t *testing.T) {
	runs := []Run{run("a", string(deno.PhasePending), 1), run("b", string(deno.PhasePending), 2)}
	groups := GroupByParent(runs, func(r Run) string {
		if r.Ref.Name == "b" {
			return ""
		}
		return "parent"
	})
	if len(groups) != 1 || len(groups["parent"]) != 1 {
		t.Fatalf("groups = %v", groups)
	}
}
