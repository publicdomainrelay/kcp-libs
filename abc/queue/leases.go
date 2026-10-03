package queue

import (
	"sync"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/deno"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

const DefaultLeaseTTL = 2 * time.Minute

type Lease struct {
	Parent ref.Ref

	Admitted time.Time
}

type Leases struct {
	TTL time.Duration

	mu sync.Mutex

	entries map[ref.Ref]Lease
}

func NewLeases(ttl time.Duration) *Leases {
	if ttl <= 0 {
		ttl = DefaultLeaseTTL
	}
	return &Leases{TTL: ttl, entries: map[ref.Ref]Lease{}}
}

func (l *Leases) Grant(run, parent ref.Ref, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries[run] = Lease{Parent: parent, Admitted: now}
}

func (l *Leases) Forget(run ref.Ref) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, run)
}

func (l *Leases) Count(parent ref.Ref, observed map[ref.Ref]string, now time.Time, isTerminal func(phase string) bool) int32 {
	terminal := Terminal(isTerminal)
	l.mu.Lock()
	defer l.mu.Unlock()
	var held int32
	for run, lease := range l.entries {
		if lease.Parent != parent {
			continue
		}
		if now.Sub(lease.Admitted) > l.TTL {
			delete(l.entries, run)
			continue
		}
		if phase, seen := observed[run]; seen && (phase == string(deno.PhaseRunning) || terminal(phase)) {
			delete(l.entries, run)
			continue
		}
		held++
	}
	return held
}

func (l *Leases) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}
