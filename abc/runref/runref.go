package runref

import (
	"sync"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

const DefaultTTL = 2 * time.Minute

type Record struct {
	Ref ref.Ref

	RunID string

	UID string

	At time.Time
}

type Current struct {
	RunID string

	UID string

	Started bool

	Retries int32
}

type Index struct {
	TTL time.Duration

	mu sync.Mutex

	entries map[string]Record
}

func New(ttl time.Duration) *Index {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Index{TTL: ttl, entries: map[string]Record{}}
}

func (i *Index) Record(r ref.Ref, runID, uid string, now time.Time) {
	if runID == "" {
		return
	}
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	i.mu.Lock()
	defer i.mu.Unlock()
	for id, entry := range i.entries {
		if now.Sub(entry.At) >= i.TTL || (id != runID && entry.Ref == key) {
			delete(i.entries, id)
		}
	}
	i.entries[runID] = Record{Ref: key, RunID: runID, UID: uid, At: now}
}

func (i *Index) Forget(r ref.Ref) {
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	i.mu.Lock()
	defer i.mu.Unlock()
	for id, entry := range i.entries {
		if entry.Ref == key {
			delete(i.entries, id)
		}
	}
}

func (i *Index) Keep(r ref.Ref, runID, uid string, terminal bool, now time.Time) {
	if terminal || runID == "" {
		i.Forget(r)
		return
	}
	i.Record(r, runID, uid, now)
}

func (i *Index) Lookup(r ref.Ref) (Record, bool) {
	key := ref.New(r.LogicalCluster, r.Namespace, r.Name)
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, entry := range i.entries {
		if entry.Ref == key {
			return entry, true
		}
	}
	return Record{}, false
}

func (i *Index) Len() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return len(i.entries)
}

func AlreadyStarted(known Record, found bool, current Current) bool {
	if !found || known.RunID == "" || known.RunID == current.RunID {
		return false
	}
	if current.Started || current.Retries > 0 {
		return false
	}
	if known.UID != "" && current.UID != "" && known.UID != current.UID {
		return false
	}
	return true
}
