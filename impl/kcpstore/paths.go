package kcpstore

import (
	"context"
	"errors"
	"net"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// transient is the difference between an answer and an outage. A namespace
// whose path cannot be read because the store said no -- there is no such
// object, the object carries no path, the caller is not allowed -- is an answer
// worth keeping. A store that was unwell at that moment is not, or one blip
// would freeze a workspace's name for the life of the process.
func transient(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return apierrors.IsInternalError(err) ||
		apierrors.IsServiceUnavailable(err) ||
		apierrors.IsServerTimeout(err) ||
		apierrors.IsTimeout(err) ||
		apierrors.IsTooManyRequests(err)
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
		if transient(err) {
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
