package joballoc

import (
	"sync"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

const DefaultTTL = 2 * time.Minute

type Allocation struct {
	Name string

	At time.Time
}

type Allocator struct {
	TTL time.Duration

	mu sync.Mutex

	byJob map[ref.Ref][]Allocation
}

func New(ttl time.Duration) *Allocator {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Allocator{TTL: ttl, byJob: map[ref.Ref][]Allocation{}}
}

func (a *Allocator) Allocate(r ref.Ref, names []string, now time.Time) {
	if len(names) == 0 {
		return
	}
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	a.mu.Lock()
	defer a.mu.Unlock()
	kept := a.byJob[key][:0]
	for _, allocation := range a.byJob[key] {
		if now.Sub(allocation.At) < a.TTL {
			kept = append(kept, allocation)
		}
	}
	for _, name := range names {
		kept = append(kept, Allocation{Name: name, At: now})
	}
	a.byJob[key] = kept
}

func (a *Allocator) Names(r ref.Ref) []string {
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.byJob[key]) == 0 {
		return nil
	}
	out := make([]string, 0, len(a.byJob[key]))
	for _, allocation := range a.byJob[key] {
		out = append(out, allocation.Name)
	}
	return out
}

func (a *Allocator) Pending(r ref.Ref, observed []string, now time.Time) []string {
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	a.mu.Lock()
	defer a.mu.Unlock()
	seen := make(map[string]bool, len(observed))
	for _, name := range observed {
		seen[name] = true
	}
	var out []string
	for _, allocation := range a.byJob[key] {
		if seen[allocation.Name] || now.Sub(allocation.At) >= a.TTL {
			continue
		}
		seen[allocation.Name] = true
		out = append(out, allocation.Name)
	}
	return out
}

func (a *Allocator) Forget(r ref.Ref) {
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.byJob, key)
}

func (a *Allocator) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.byJob)
}

func MergeNames(groups ...[]string) []string {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	seen := make(map[string]bool, total)
	out := make([]string, 0, total)
	for _, group := range groups {
		for _, name := range group {
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}
