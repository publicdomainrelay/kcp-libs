package kcpstore

import (
	"context"
	"errors"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// definitive is the store answering, as opposed to the store failing. Only an
// answer belongs in the cache: there is no such workspace, it carries no path,
// the caller may not read it. Everything else -- a gateway that was down, a
// body that did not parse, a call someone cancelled -- is asked again, because
// one bad response must not freeze a workspace's name for the life of the
// process. The list is an allowlist rather than a list of failures to retry, so
// an error nobody thought of is retried rather than cached.
func definitive(err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, ErrNoPath):
		return true
	case IsNotFound(err), apierrors.IsForbidden(err), apierrors.IsUnauthorized(err):
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
