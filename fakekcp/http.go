package fakekcp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	statusNotFound = "NotFound"

	statusConflict = "Conflict"

	statusAlreadyExists = "AlreadyExists"

	statusBadRequest = "BadRequest"
)

func (c *Cluster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segments := splitPath(r.URL.Path)
	cluster, rest, ok := splitCluster(segments)
	if !ok {
		writeStatus(w, http.StatusNotFound, statusNotFound, "no /clusters/<name> segment in "+r.URL.Path)
		return
	}
	switch {
	case isEndpointSlices(rest):
		c.serveEndpointSlices(w, r)
		return
	case isLogicalCluster(rest):
		c.serveLogicalCluster(w, r, cluster)
		return
	case isTokenRequest(rest):
		c.serveToken(w, r, rest)
		return
	}
	group, version, scope, resource, name, sub, ok := splitResource(rest)
	if !ok {
		writeStatus(w, http.StatusNotFound, statusNotFound, "no resource in "+r.URL.Path)
		return
	}
	c.serveResource(w, r, cluster, group, version, scope, resource, name, sub)
}

func (c *Cluster) serveResource(w http.ResponseWriter, r *http.Request, cluster, group, version, namespace, resource, name, sub string) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("watch") == "true" {
			c.serveWatch(w, r, cluster, group, version, resource)
			return
		}
		if name == "" {
			c.serveList(w, cluster, group, version, resource)
			return
		}
		body, found := c.Get(cluster, namespace, resource, name)
		if !found {
			writeStatus(w, http.StatusNotFound, statusNotFound, resource+" "+name+" not found")
			return
		}
		writeJSON(w, http.StatusOK, body)
	case http.MethodPost:
		body, err := readObject(r)
		if err != nil {
			writeStatus(w, http.StatusBadRequest, statusBadRequest, err.Error())
			return
		}
		if _, exists := c.Get(cluster, namespace, resource, NameOf(body)); exists {
			writeStatus(w, http.StatusConflict, statusAlreadyExists, resource+" "+NameOf(body)+" already exists")
			return
		}
		writeJSON(w, http.StatusCreated, c.Create(cluster, namespace, resource, body))
	case http.MethodPatch:
		c.servePatch(w, r, cluster, namespace, resource, name, sub)
	case http.MethodDelete:
		if !c.Delete(cluster, namespace, resource, name) {
			writeStatus(w, http.StatusNotFound, statusNotFound, resource+" "+name+" not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"kind": "Status", "status": "Success"})
	default:
		writeStatus(w, http.StatusMethodNotAllowed, statusBadRequest, r.Method+" is not supported")
	}
}

func (c *Cluster) serveList(w http.ResponseWriter, cluster, group, version, resource string) {
	items := c.List(cluster, resource)
	if items == nil {
		items = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"apiVersion": versionOf(group, version),
		"kind":       kindOf(resource) + "List",
		"metadata":   map[string]any{"resourceVersion": c.resourceVersion()},
		"items":      items,
	})
}

func (c *Cluster) serveWatch(w http.ResponseWriter, r *http.Request, cluster, group, version, resource string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeStatus(w, http.StatusInternalServerError, statusBadRequest, "the response cannot stream")
		return
	}
	id, watcher := c.addWatcher(cluster, resource)
	defer c.removeWatcher(id)

	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)

	if r.URL.Query().Get("sendInitialEvents") == "true" {
		for _, object := range c.watchObjects(cluster, resource) {
			if err := encoder.Encode(watchFrame("ADDED", object)); err != nil {
				return
			}
		}
		if err := encoder.Encode(bookmark(versionOf(group, version), kindOf(resource), c.resourceVersion())); err != nil {
			return
		}
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-watcher.done:
			return
		case frame := <-watcher.events:
			if err := encoder.Encode(frame); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (c *Cluster) servePatch(w http.ResponseWriter, r *http.Request, cluster, namespace, resource, name, _ string) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeStatus(w, http.StatusBadRequest, statusBadRequest, err.Error())
		return
	}
	body, found := c.Get(cluster, namespace, resource, name)
	if !found {
		writeStatus(w, http.StatusNotFound, statusNotFound, resource+" "+name+" not found")
		return
	}
	c.mu.Lock()
	c.patches++
	c.mu.Unlock()

	if strings.Contains(r.Header.Get("Content-Type"), "json-patch") {
		patched, err := applyJSONPatch(body, raw)
		if err != nil {
			writeStatus(w, http.StatusConflict, statusConflict, err.Error())
			return
		}
		c.Apply(cluster, namespace, resource, patched)
		writeJSON(w, http.StatusOK, patched)
		return
	}
	patch, err := decodeObject(raw)
	if err != nil {
		writeStatus(w, http.StatusBadRequest, statusBadRequest, err.Error())
		return
	}
	for key, value := range patch {
		if key == "metadata" {
			continue
		}
		body[key] = mergeValue(body[key], value)
	}
	c.Apply(cluster, namespace, resource, body)
	writeJSON(w, http.StatusOK, body)
}

func (c *Cluster) serveEndpointSlices(w http.ResponseWriter, _ *http.Request) {
	c.mu.Lock()
	slices := append([]endpointSlice(nil), c.slices...)
	c.mu.Unlock()
	items := make([]map[string]any, 0, len(slices))
	for _, slice := range slices {
		items = append(items, map[string]any{
			"apiVersion": "apis.kcp.io/v1alpha1",
			"kind":       "APIExportEndpointSlice",
			"metadata":   map[string]any{"name": slice.export + "-slice"},
			"spec":       map[string]any{"export": map[string]any{"name": slice.export}},
			"status":     map[string]any{"endpoints": []any{map[string]any{"url": slice.url}}},
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"apiVersion": "apis.kcp.io/v1alpha1",
		"kind":       "APIExportEndpointSliceList",
		"metadata":   map[string]any{"resourceVersion": c.resourceVersion()},
		"items":      items,
	})
}

func (c *Cluster) serveLogicalCluster(w http.ResponseWriter, _ *http.Request, cluster string) {
	c.mu.Lock()
	path := c.paths[cluster]
	c.mu.Unlock()
	if path == "" {
		writeStatus(w, http.StatusNotFound, statusNotFound, "logical cluster "+cluster+" is unknown")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"apiVersion": "core.kcp.io/v1alpha1",
		"kind":       "LogicalCluster",
		"metadata": map[string]any{
			"name":        "cluster",
			"annotations": map[string]any{PathAnnotation: path},
		},
	})
}

func (c *Cluster) serveToken(w http.ResponseWriter, r *http.Request, rest []string) {
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed, statusBadRequest, r.Method+" is not supported")
		return
	}
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()
	writeJSON(w, http.StatusCreated, map[string]any{
		"apiVersion": "authentication.k8s.io/v1",
		"kind":       "TokenRequest",
		"metadata":   map[string]any{"name": rest[len(rest)-2]},
		"status":     map[string]any{"token": token, "expirationTimestamp": time.Now().Add(time.Hour).UTC().Format(time.RFC3339)},
	})
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func splitCluster(segments []string) (string, []string, bool) {
	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == "clusters" {
			return segments[i+1], segments[i+2:], true
		}
	}
	return "", nil, false
}

func isEndpointSlices(rest []string) bool {
	return len(rest) == 4 && rest[0] == "apis" && rest[1] == "apis.kcp.io" && rest[3] == "apiexportendpointslices"
}

func isLogicalCluster(rest []string) bool {
	return len(rest) == 5 && rest[0] == "apis" && rest[1] == "core.kcp.io" && rest[3] == "logicalclusters"
}

func isTokenRequest(rest []string) bool {
	return len(rest) == 7 && rest[0] == "api" && rest[1] == "v1" && rest[2] == "namespaces" && rest[4] == "serviceaccounts" && rest[6] == "token"
}

func splitResource(rest []string) (group, version, namespace, resource, name, sub string, ok bool) {
	if len(rest) < 3 || (rest[0] != "apis" && rest[0] != "api") {
		return "", "", "", "", "", "", false
	}
	if rest[0] == "api" {
		group = ""
		version = rest[1]
		rest = rest[2:]
	} else {
		group = rest[1]
		version = rest[2]
		rest = rest[3:]
	}
	if len(rest) >= 3 && rest[0] == "namespaces" {
		namespace = rest[1]
		rest = rest[2:]
	}
	if len(rest) == 0 {
		return "", "", "", "", "", "", false
	}
	resource = rest[0]
	rest = rest[1:]
	if len(rest) > 0 {
		name = rest[0]
		rest = rest[1:]
	}
	if len(rest) > 0 {
		sub = rest[0]
	}
	return group, version, namespace, resource, name, sub, true
}

func versionOf(group, version string) string {
	if group == "" {
		return version
	}
	return group + "/" + version
}

func kindOf(resource string) string {
	if resource == "" {
		return ""
	}
	return strings.ToUpper(resource[:1]) + resource[1:]
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeStatus(w http.ResponseWriter, status int, reason, message string) {
	writeJSON(w, status, map[string]any{
		"kind":       "Status",
		"apiVersion": "v1",
		"status":     "Failure",
		"reason":     reason,
		"message":    message,
		"code":       status,
	})
}

func decodeObject(raw []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func readObject(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return decodeObject(raw)
}

func mergeValue(current, patch any) any {
	patchMap, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	currentMap, ok := current.(map[string]any)
	if !ok {
		currentMap = map[string]any{}
	}
	out := map[string]any{}
	for key, value := range currentMap {
		out[key] = value
	}
	for key, value := range patchMap {
		if value == nil {
			delete(out, key)
			continue
		}
		out[key] = mergeValue(out[key], value)
	}
	return out
}

type jsonPatchOp struct {
	Op string `json:"op"`

	Path string `json:"path"`

	Value any `json:"value"`
}

func applyJSONPatch(body map[string]any, raw []byte) (map[string]any, error) {
	var ops []jsonPatchOp
	if err := json.Unmarshal(raw, &ops); err != nil {
		return nil, err
	}
	out := clone(body)
	for _, op := range ops {
		segments := splitPath(op.Path)
		if len(segments) == 0 {
			return nil, errBadPatchPath(op.Path)
		}
		target := out
		for _, segment := range segments[:len(segments)-1] {
			target = ensureMap(target, segment)
		}
		last := segments[len(segments)-1]
		switch op.Op {
		case "test":
			if !valuesEqual(target[last], op.Value) {
				return nil, errPatchTestFailed(op.Path)
			}
		case "add", "replace":
			target[last] = cloneValue(op.Value)
		case "remove":
			delete(target, last)
		default:
			return nil, errBadPatchPath(op.Op)
		}
	}
	return out, nil
}

func valuesEqual(a, b any) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(left) == string(right)
}

type patchError string

func (e patchError) Error() string {
	return string(e)
}

func errBadPatchPath(path string) error {
	return patchError("unsupported patch path " + path)
}

func errPatchTestFailed(path string) error {
	return patchError("test failed for " + path)
}
