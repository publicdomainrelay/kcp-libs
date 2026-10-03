package queue

import (
	"fmt"
	"sort"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/deno"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Policy string

const (
	PolicyAllow Policy = Policy(deno.ConcurrencyAllow)

	PolicyForbid Policy = Policy(deno.ConcurrencyForbid)

	PolicyReplace Policy = Policy(deno.ConcurrencyReplace)
)

type Run struct {
	Ref ref.Ref

	Phase string

	Created time.Time
}

type Capacity struct {
	Policy Policy

	MaxConcurrent *int32
}

type Blocker struct {
	Reason string

	Message string
}

type Admission struct {
	Gated bool

	Allowed bool

	Position int32

	Active int32

	Limit int32

	Unlimited bool

	Reason string

	Message string

	Preempt []ref.Ref
}

func Limit(policy Policy, maxConcurrent *int32) (int32, bool) {
	if policy == PolicyAllow {
		if maxConcurrent == nil || *maxConcurrent <= 0 {
			return 0, true
		}
		return *maxConcurrent, false
	}
	return 1, false
}

func Decision(policy Policy, maxConcurrent *int32, active, ahead int32) (bool, string, string) {
	limit, unlimited := Limit(policy, maxConcurrent)
	if unlimited || active+ahead < limit {
		return true, "", ""
	}
	display := policy
	if display == "" {
		display = PolicyForbid
	}
	return false, deno.ReasonAtCapacity,
		fmt.Sprintf("waiting for a free slot (concurrencyPolicy=%s, maxConcurrent=%d)", display, limit)
}

func Order(runs []Run) {
	sort.SliceStable(runs, func(a, b int) bool {
		if !runs[a].Created.Equal(runs[b].Created) {
			return runs[a].Created.Before(runs[b].Created)
		}
		return runs[a].Ref.Name < runs[b].Ref.Name
	})
}

func Plan(runs []Run, capacity Capacity, blocker *Blocker, reserved int32, isTerminal func(phase string) bool) []Admission {
	terminal := Terminal(isTerminal)
	ordered := append([]Run(nil), runs...)
	Order(ordered)

	active := reserved
	var pending []int
	for i := range ordered {
		switch {
		case terminal(ordered[i].Phase):
		case ordered[i].Phase == string(deno.PhaseRunning):
			active++
		default:
			pending = append(pending, i)
		}
	}

	limit, unlimited := Limit(capacity.Policy, capacity.MaxConcurrent)

	newestPending := -1
	if capacity.Policy == PolicyReplace && len(pending) > 0 {
		newestPending = pending[len(pending)-1]
	}

	out := make([]Admission, len(ordered))
	for i := range ordered {
		out[i] = Admission{Gated: false, Allowed: true, Active: active, Limit: limit, Unlimited: unlimited}
	}

	for position, i := range pending {
		admission := Admission{
			Gated:     true,
			Position:  int32(position),
			Active:    active,
			Limit:     limit,
			Unlimited: unlimited,
		}
		switch {
		case blocker != nil:
			admission.Reason = blocker.Reason
			admission.Message = blocker.Message
		case capacity.Policy == PolicyReplace:
			if i == newestPending {
				admission.Allowed = true
				for j := range ordered {
					if j == i || terminal(ordered[j].Phase) {
						continue
					}
					admission.Preempt = append(admission.Preempt, ordered[j].Ref)
				}
			} else {
				admission.Reason = deno.ReasonSuperseded
				admission.Message = "a newer run supersedes this one under concurrencyPolicy=Replace"
			}
		default:
			admission.Allowed, admission.Reason, admission.Message = Decision(capacity.Policy, capacity.MaxConcurrent, active, int32(position))
		}
		out[i] = admission
	}
	return out
}

func PlanIndex(runs []Run, capacity Capacity, blocker *Blocker, reserved int32, isTerminal func(phase string) bool) map[ref.Ref]Admission {
	ordered := append([]Run(nil), runs...)
	Order(ordered)
	planned := Plan(ordered, capacity, blocker, reserved, isTerminal)
	out := make(map[ref.Ref]Admission, len(ordered))
	for i := range ordered {
		out[ordered[i].Ref] = planned[i]
	}
	return out
}

func PlanByParent(runs []Run, parent func(Run) string, capacity func(parent string) Capacity, blocker func(parent string) *Blocker, reserved map[string]int32, isTerminal func(phase string) bool) map[ref.Ref]Admission {
	groups := GroupByParent(runs, parent)
	out := make(map[ref.Ref]Admission, len(runs))
	for name, group := range groups {
		planned := Plan(group, capacity(name), blocker(name), reserved[name], isTerminal)
		for i := range group {
			out[group[i].Ref] = planned[i]
		}
	}
	return out
}

func GroupByParent(runs []Run, parent func(Run) string) map[string][]Run {
	groups := map[string][]Run{}
	order := []string{}
	for _, run := range runs {
		name := parent(run)
		if name == "" {
			continue
		}
		if _, seen := groups[name]; !seen {
			order = append(order, name)
		}
		groups[name] = append(groups[name], run)
	}
	return groups
}

func WakeList(runs []Run, isTerminal func(phase string) bool, limit int32, unlimited bool) []ref.Ref {
	terminal := Terminal(isTerminal)
	var pending []Run
	for _, run := range runs {
		if run.Phase == string(deno.PhaseRunning) || terminal(run.Phase) {
			continue
		}
		pending = append(pending, run)
	}
	Order(pending)
	count := len(pending)
	if !unlimited && int(limit) < count {
		count = int(limit)
	}
	out := make([]ref.Ref, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, pending[i].Ref)
	}
	return out
}

func Terminal(isTerminal func(phase string) bool) func(phase string) bool {
	if isTerminal != nil {
		return isTerminal
	}
	return deno.TerminalPolicyWorkflow
}
