package kcpstore

import (
	"context"
	"errors"
	"sync"
)

// definitive is the store's answer, as opposed to the store failing. Only an
// answer that cannot change by asking again belongs in the cache: there is no
// such workspace, or it exists and carries no path. Everything else is asked
// again -- a gateway that was down, a body that did not parse, a call someone
// cancelled, and a refusal, because a token mid-refresh and a grant not yet
// visible are refusals that heal. The list is an allowlist, so an error nobody
// thought of is retried rather than frozen into a workspace's name.
func definitive(err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, ErrNoPath):
		return true
	case IsNotFound(err):
		return true
	}
	return false
}

type PathCache struct {
	store *Store

	mu sync.Mutex

	byID map[string]string
}

func NewPathCache(store *Store) *PathCache {
	return &PathCache{store: store, byID: map[string]string{}}
}

func (c *PathCache) Lookup(ctx context.Context, logicalCluster string) string {
	c.mu.Lock()
	if path, seen := c.byID[logicalCluster]; seen {
		c.mu.Unlock()
		return path
	}
	c.mu.Unlock()

	path := ""
	if c.store != nil {
		resolved, err := c.store.ClusterPath(ctx, logicalCluster)
		if !definitive(err) {
			return ""
		}
		path = resolved
	}
	c.mu.Lock()
	c.byID[logicalCluster] = path
	c.mu.Unlock()
	return path
}

func (c *PathCache) Forget(logicalCluster string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.byID, logicalCluster)
}

func (c *PathCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.byID)
}
