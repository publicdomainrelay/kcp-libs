package fakekcp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	jsonContentType = "application/json"

	ClusterAnnotation = "kcp.io/cluster"

	PathAnnotation = "kcp.io/path"

	Wildcard = "*"
)

type Event struct {
	Type string

	Cluster string

	Namespace string

	Resource string

	Name string

	Object map[string]any

	At time.Time
}

type stored struct {
	cluster string

	namespace string

	resource string

	name string

	body map[string]any

	version uint64
}

type endpointSlice struct {
	export string

	url string
}

type watcher struct {
	cluster string

	resource string

	events chan map[string]any

	done chan struct{}
}

type Cluster struct {
	mu sync.Mutex

	version uint64

	objects map[string]*stored

	order []string

	paths map[string]string

	slices []endpointSlice

	watchers map[int]*watcher

	nextWatcher int

	events []Event

	patches int

	token string

	listener net.Listener

	server *http.Server

	url string
}

func New() (*Cluster, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("fakekcp: listen: %w", err)
	}
	cluster := &Cluster{
		objects:  map[string]*stored{},
		paths:    map[string]string{},
		watchers: map[int]*watcher{},
		token:    "fake-service-account-token",
		listener: listener,
		url:      "http://" + listener.Addr().String(),
	}
	cluster.server = &http.Server{Handler: cluster}
	go func() {
		_ = cluster.server.Serve(listener)
	}()
	return cluster, nil
}

func (c *Cluster) URL() string {
	return c.url
}

func (c *Cluster) Close() error {
	c.mu.Lock()
	for id, w := range c.watchers {
		close(w.done)
		delete(c.watchers, id)
	}
	c.mu.Unlock()
	return c.server.Close()
}

func (c *Cluster) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

func (c *Cluster) SetClusterPath(id, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths[id] = path
}

func (c *Cluster) AddEndpointSlice(export, url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.slices = append(c.slices, endpointSlice{export: export, url: url})
}

func (c *Cluster) Create(cluster, namespace, resource string, body map[string]any) map[string]any {
	return c.put(cluster, namespace, resource, body, "ADDED")
}

func (c *Cluster) Apply(cluster, namespace, resource string, body map[string]any) map[string]any {
	name := NameOf(body)
	_, exists := c.Get(cluster, namespace, resource, name)
	event := "ADDED"
	if exists {
		event = "MODIFIED"
	}
	return c.put(cluster, namespace, resource, body, event)
}

func (c *Cluster) put(cluster, namespace, resource string, body map[string]any, event string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	name := NameOf(body)
	key := objectKey(cluster, namespace, resource, name)
	next := clone(body)
	metadata := ensureMap(next, "metadata")
	if metadata["namespace"] == nil && namespace != "" {
		metadata["namespace"] = namespace
	}
	if metadata["name"] == nil {
		metadata["name"] = name
	}
	if metadata["uid"] == nil {
		metadata["uid"] = fmt.Sprintf("uid-%s-%s", resource, name)
	}
	if metadata["creationTimestamp"] == nil {
		metadata["creationTimestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if metadata["generation"] == nil {
		metadata["generation"] = 1
	}
	c.version++
	version := c.version
	metadata["resourceVersion"] = strconv.FormatUint(version, 10)
	if _, seen := c.objects[key]; !seen {
		c.order = append(c.order, key)
	}
	c.objects[key] = &stored{
		cluster:   cluster,
		namespace: namespace,
		resource:  resource,
		name:      name,
		body:      next,
		version:   version,
	}
	c.broadcastLocked(event, c.objects[key])
	return clone(stripAnnotations(next))
}

func (c *Cluster) Get(cluster, namespace, resource, name string) (map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.objects[objectKey(cluster, namespace, resource, name)]
	if !ok {
		return nil, false
	}
	return clone(stripAnnotations(entry.body)), true
}

func (c *Cluster) List(cluster, resource string) []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []map[string]any
	for _, key := range c.order {
		entry := c.objects[key]
		if entry == nil || entry.resource != resource {
			continue
		}
		if cluster != Wildcard && entry.cluster != cluster {
			continue
		}
		out = append(out, c.servedLocked(entry, cluster == Wildcard))
	}
	return out
}

func (c *Cluster) SetStatus(cluster, namespace, resource, name string, status map[string]any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.objects[objectKey(cluster, namespace, resource, name)]
	if !ok {
		return false
	}
	entry.body["status"] = clone(status)
	c.version++
	entry.version = c.version
	ensureMap(entry.body, "metadata")["resourceVersion"] = strconv.FormatUint(c.version, 10)
	c.broadcastLocked("MODIFIED", entry)
	return true
}

func (c *Cluster) Delete(cluster, namespace, resource, name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := objectKey(cluster, namespace, resource, name)
	entry, ok := c.objects[key]
	if !ok {
		return false
	}
	delete(c.objects, key)
	for i, candidate := range c.order {
		if candidate == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.broadcastLocked("DELETED", entry)
	return true
}

func (c *Cluster) Events() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Event(nil), c.events...)
}

func (c *Cluster) Patches() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.patches
}

func (c *Cluster) WaitFor(ctx context.Context, condition func() bool) error {
	deadline, hasDeadline := ctx.Deadline()
	for {
		if condition() {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if hasDeadline && time.Now().After(deadline) {
			return fmt.Errorf("fakekcp: timed out waiting for the condition")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
}

func (c *Cluster) recordLocked(event, resource string, entry *stored, body map[string]any) {
	c.events = append(c.events, Event{
		Type:      event,
		Cluster:   entry.cluster,
		Namespace: entry.namespace,
		Resource:  resource,
		Name:      entry.name,
		Object:    body,
		At:        time.Now(),
	})
}

func (c *Cluster) broadcastLocked(event string, entry *stored) {
	c.recordLocked(event, entry.resource, entry, clone(stripAnnotations(entry.body)))
	for _, w := range c.watchers {
		if w.resource != entry.resource {
			continue
		}
		if w.cluster != Wildcard && w.cluster != entry.cluster {
			continue
		}
		frame := watchFrame(event, c.servedLocked(entry, w.cluster == Wildcard))
		select {
		case w.events <- frame:
		case <-w.done:
		default:
		}
	}
}

func (c *Cluster) servedLocked(entry *stored, stampCluster bool) map[string]any {
	body := clone(entry.body)
	if stampCluster {
		ensureMap(ensureMap(body, "metadata"), "annotations")[ClusterAnnotation] = entry.cluster
	}
	return body
}

func (c *Cluster) addWatcher(cluster, resource string) (int, *watcher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextWatcher++
	id := c.nextWatcher
	w := &watcher{
		cluster:  cluster,
		resource: resource,
		events:   make(chan map[string]any, 256),
		done:     make(chan struct{}),
	}
	c.watchers[id] = w
	return id, w
}

func (c *Cluster) removeWatcher(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w, ok := c.watchers[id]; ok {
		close(w.done)
		delete(c.watchers, id)
	}
}

func (c *Cluster) watchObjects(cluster, resource string) []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []map[string]any
	for _, key := range c.order {
		entry := c.objects[key]
		if entry == nil || entry.resource != resource {
			continue
		}
		if cluster != Wildcard && entry.cluster != cluster {
			continue
		}
		out = append(out, c.servedLocked(entry, cluster == Wildcard))
	}
	return out
}

func (c *Cluster) resourceVersion() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strconv.FormatUint(c.version, 10)
}

func objectKey(cluster, namespace, resource, name string) string {
	return cluster + "|" + resource + "|" + namespace + "|" + name
}
