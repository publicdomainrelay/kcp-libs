package fakekcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	group    = "example.computer"
	version  = "v1alpha1"
	resource = "widgets"
)

func get(t *testing.T, cluster *Cluster, path string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(cluster.URL() + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(body, &out)
	return resp.StatusCode, out
}

func do(t *testing.T, method, url, contentType string, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func TestCreateGetAndList(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()

	cluster.Create("root:alice", "default", resource, Object(group+"/"+version, "Widget", "default", "one"))
	cluster.Create("root:bob", "default", resource, Object(group+"/"+version, "Widget", "default", "two"))

	if body, found := cluster.Get("root:alice", "default", resource, "one"); !found || PhaseOf(body) != "" {
		t.Fatalf("get = (%v, %v)", body, found)
	}
	if _, found := cluster.Get("root:alice", "default", resource, "two"); found {
		t.Fatal("an object belongs to one cluster")
	}

	status, listed := get(t, cluster, "/clusters/root:alice/apis/"+group+"/"+version+"/namespaces/default/"+resource)
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	items, _ := listed["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items = %v", items)
	}

	status, wildcard := get(t, cluster, "/clusters/*/apis/"+group+"/"+version+"/"+resource)
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	items, _ = wildcard["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("wildcard items = %v", items)
	}
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		metadata, _ := item["metadata"].(map[string]any)
		annotations, _ := metadata["annotations"].(map[string]any)
		if annotations[ClusterAnnotation] == "" {
			t.Fatalf("a wildcard list must stamp the cluster: %v", item)
		}
	}
	if metadata, _ := wildcard["metadata"].(map[string]any); metadata["resourceVersion"] == nil {
		t.Fatal("a list must carry a resource version")
	}
}

func TestStatusMergePatchAndFinalizerJSONPatch(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	base := cluster.URL() + "/clusters/root:alice/apis/" + group + "/" + version + "/namespaces/default/" + resource + "/one"
	cluster.Create("root:alice", "default", resource, WithLabel(Object(group+"/"+version, "Widget", "default", "one"), "a", "b"))

	status, _ := do(t, http.MethodPatch, base+"/status", "application/merge-patch+json", `{"status":{"phase":"Running","observed":2}}`)
	if status != http.StatusOK {
		t.Fatalf("status patch = %d", status)
	}
	body, _ := cluster.Get("root:alice", "default", resource, "one")
	if PhaseOf(body) != "Running" {
		t.Fatalf("phase = %q", PhaseOf(body))
	}
	if _, ok := body["spec"]; !ok {
		t.Fatal("a status patch must not drop the rest of the object")
	}

	patch := `[{"op":"test","path":"/metadata/finalizers","value":["f"]},{"op":"add","path":"/metadata/finalizers","value":[]}]`
	endpoint := cluster.URL() + "/clusters/root:alice/apis/" + group + "/" + version + "/namespaces/default/" + resource + "/one"
	cluster.Apply("root:alice", "default", resource, map[string]any{
		"metadata": map[string]any{"name": "one", "namespace": "default", "finalizers": []any{"f"}},
	})
	if status, _ := do(t, http.MethodPatch, endpoint, "application/json-patch+json", patch); status != http.StatusOK {
		t.Fatalf("finalizer patch = %d", status)
	}
	if status, _ := do(t, http.MethodPatch, endpoint, "application/json-patch+json", patch); status != http.StatusConflict {
		t.Fatalf("a failing test op must conflict, got %d", status)
	}
	if cluster.Patches() != 3 {
		t.Fatalf("patches = %d", cluster.Patches())
	}
}

func TestCreateConflictDeleteAndNotFound(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	endpoint := cluster.URL() + "/clusters/root:alice/apis/" + group + "/" + version + "/namespaces/default/" + resource
	object := `{"apiVersion":"` + group + `/` + version + `","kind":"Widget","metadata":{"name":"one","namespace":"default"}}`

	if status, _ := do(t, http.MethodPost, endpoint, "application/json", object); status != http.StatusCreated {
		t.Fatalf("create = %d", status)
	}
	if status, _ := do(t, http.MethodPost, endpoint, "application/json", object); status != http.StatusConflict {
		t.Fatalf("a repeated create must conflict, got %d", status)
	}
	if status, _ := do(t, http.MethodDelete, endpoint+"/one", "", ""); status != http.StatusOK {
		t.Fatalf("delete = %d", status)
	}
	if status, body := do(t, http.MethodDelete, endpoint+"/one", "", ""); status != http.StatusNotFound {
		t.Fatalf("delete again = %d %v", status, body)
	}
	if status, body := get(t, cluster, "/clusters/root:alice/apis/"+group+"/"+version+"/namespaces/default/"+resource+"/one"); status != http.StatusNotFound {
		t.Fatalf("get after delete = %d %v", status, body)
	}
	if body, _ := json.Marshal(cluster.Events()); !bytes.Contains(body, []byte("DELETED")) {
		t.Fatal("a delete must be recorded as an event")
	}
}

func TestWatchStreamsFrames(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cluster.URL()+"/clusters/*/apis/"+group+"/"+version+"/"+resource+"?watch=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	frames := make(chan map[string]any, 4)
	go func() {
		decoder := json.NewDecoder(resp.Body)
		for {
			var frame map[string]any
			if err := decoder.Decode(&frame); err != nil {
				return
			}
			frames <- frame
		}
	}()

	cluster.Create("root:alice", "default", resource, Object(group+"/"+version, "Widget", "default", "watched"))
	select {
	case frame := <-frames:
		if frame["type"] != "ADDED" {
			t.Fatalf("frame = %v", frame)
		}
		object, _ := frame["object"].(map[string]any)
		metadata, _ := object["metadata"].(map[string]any)
		annotations, _ := metadata["annotations"].(map[string]any)
		if annotations[ClusterAnnotation] != "root:alice" {
			t.Fatalf("a wildcard watch must stamp the cluster: %v", object)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no watch frame arrived")
	}
}

func TestWatchInitialEventsEndWithABookmark(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	cluster.Create("root:alice", "default", resource, Object(group+"/"+version, "Widget", "default", "seeded"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := cluster.URL() + "/clusters/*/apis/" + group + "/" + version + "/" + resource + "?watch=true&sendInitialEvents=true&resourceVersionMatch=NotOlderThan"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var frame map[string]any
	if err := decoder.Decode(&frame); err != nil {
		t.Fatal(err)
	}
	if frame["type"] != "ADDED" {
		t.Fatalf("first frame = %v", frame)
	}
	if err := decoder.Decode(&frame); err != nil {
		t.Fatal(err)
	}
	if frame["type"] != "BOOKMARK" {
		t.Fatalf("second frame = %v", frame)
	}
	object, _ := frame["object"].(map[string]any)
	metadata, _ := object["metadata"].(map[string]any)
	annotations, _ := metadata["annotations"].(map[string]any)
	if annotations["k8s.io/initial-events-end"] != "true" {
		t.Fatalf("bookmark = %v", object)
	}
}

func TestEndpointSlicesClusterPathAndToken(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	cluster.SetClusterPath("2j35", "root:alice")
	cluster.AddEndpointSlice("denoruntime", "https://kcp.example/services/apiexport/root:p/denoruntime")
	cluster.SetToken("minted")

	status, slices := get(t, cluster, "/clusters/root:p/apis/apis.kcp.io/v1alpha1/apiexportendpointslices")
	if status != http.StatusOK {
		t.Fatalf("slices = %d", status)
	}
	items, _ := slices["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items = %v", items)
	}

	status, logical := get(t, cluster, "/clusters/2j35/apis/core.kcp.io/v1alpha1/logicalclusters/cluster")
	if status != http.StatusOK {
		t.Fatalf("logical cluster = %d", status)
	}
	metadata, _ := logical["metadata"].(map[string]any)
	annotations, _ := metadata["annotations"].(map[string]any)
	if annotations[PathAnnotation] != "root:alice" {
		t.Fatalf("path = %v", annotations)
	}
	if status, _ := get(t, cluster, "/clusters/missing/apis/core.kcp.io/v1alpha1/logicalclusters/cluster"); status != http.StatusNotFound {
		t.Fatalf("an unknown cluster must be absent, got %d", status)
	}

	status, token := do(t, http.MethodPost, cluster.URL()+"/clusters/2j35/api/v1/namespaces/default/serviceaccounts/reader/token", "application/json", "{}")
	if status != http.StatusCreated {
		t.Fatalf("token = %d", status)
	}
	requestStatus, _ := token["status"].(map[string]any)
	if requestStatus["token"] != "minted" {
		t.Fatalf("token = %v", token)
	}
}

func TestWaitForAndEvents(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cluster.Create("root:alice", "default", resource, Object(group+"/"+version, "Widget", "default", "late"))
	}()
	if err := cluster.WaitFor(ctx, func() bool { return len(cluster.Events()) > 0 }); err != nil {
		t.Fatal(err)
	}
	events := cluster.Events()
	if len(events) != 1 || events[0].Type != "ADDED" || events[0].Name != "late" {
		t.Fatalf("events = %+v", events)
	}

	short, cancelShort := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelShort()
	if err := cluster.WaitFor(short, func() bool { return false }); err == nil {
		t.Fatal("a condition that never holds must time out")
	}
}

func TestStateIsClonedOnTheWayOut(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	cluster.Create("root:alice", "default", resource, Object(group+"/"+version, "Widget", "default", "one"))

	body, _ := cluster.Get("root:alice", "default", resource, "one")
	body["status"].(map[string]any)["phase"] = "tampered"
	again, _ := cluster.Get("root:alice", "default", resource, "one")
	if PhaseOf(again) == "tampered" {
		t.Fatal("a caller must not be able to mutate the stored object")
	}
}

func TestServeHTTPRejectsAnUnknownPath(t *testing.T) {
	cluster, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()
	if status, _ := get(t, cluster, "/nope"); status != http.StatusNotFound {
		t.Fatalf("status = %d", status)
	}
	if status, _ := get(t, cluster, "/clusters/root:alice/apis"); status != http.StatusNotFound {
		t.Fatalf("status = %d", status)
	}
	if status, _ := get(t, cluster, "/clusters/root:alice/apis/"+group+"/"+version); status != http.StatusNotFound {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(cluster.URL(), "127.0.0.1:") {
		t.Fatalf("url = %q", cluster.URL())
	}
}
