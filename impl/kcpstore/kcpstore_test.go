package kcpstore

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/publicdomainrelay/kcp-libs/abc/store"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type runObject struct {
	Metadata struct {
		Name            string   `json:"name"`
		Namespace       string   `json:"namespace"`
		Finalizers      []string `json:"finalizers"`
		ResourceVersion string   `json:"resourceVersion"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

var denoRuns = schema.GroupVersionResource{Group: "deno.computer", Version: "v1alpha1", Resource: "denoruns"}

var _ store.Resource[runObject] = Of[runObject](nil, denoRuns)

type request struct {
	Method string

	Path string

	ContentType string

	Body string

	Authorization string
}

func newServer(t *testing.T, want []request, respond func(w http.ResponseWriter, r *http.Request, index int)) (*Store, *[]request) {
	t.Helper()
	seen := &[]request{}
	index := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*seen = append(*seen, request{
			Method:        r.Method,
			Path:          r.URL.Path,
			ContentType:   r.Header.Get("Content-Type"),
			Body:          string(body),
			Authorization: r.Header.Get("Authorization"),
		})
		respond(w, r, index)
		index++
	}))
	t.Cleanup(server.Close)
	store, err := New(Options{Host: server.URL, RestConfig: nil})
	if err != nil {
		t.Fatal(err)
	}
	if want != nil {
		t.Cleanup(func() {
			if len(*seen) != len(want) {
				t.Fatalf("requests = %d, want %d: %+v", len(*seen), len(want), *seen)
			}
			for i := range want {
				if (*seen)[i].Method != want[i].Method || (*seen)[i].Path != want[i].Path {
					t.Fatalf("request %d = %s %s, want %s %s", i, (*seen)[i].Method, (*seen)[i].Path, want[i].Method, want[i].Path)
				}
			}
		})
	}
	return store, seen
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestGetDecodesThroughTheClusterPath(t *testing.T) {
	store, _ := newServer(t, []request{
		{Method: "GET", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns/run-1"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{
			"metadata": map[string]any{"name": "run-1", "namespace": "default"},
			"status":   map[string]any{"phase": "Running"},
		})
	})
	resource := Of[runObject](store, denoRuns)
	found, err := resource.Get(context.Background(), ref.New("root:alice", "default", "run-1"))
	if err != nil {
		t.Fatal(err)
	}
	if found.Status.Phase != "Running" || found.Metadata.Name != "run-1" {
		t.Fatalf("run = %+v", found)
	}
}

func TestListDecodesItems(t *testing.T) {
	store, _ := newServer(t, []request{
		{Method: "GET", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/denoruns"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{"items": []any{
			map[string]any{"metadata": map[string]any{"name": "a"}},
			map[string]any{"metadata": map[string]any{"name": "b"}},
		}})
	})
	resource := Of[runObject](store, denoRuns)
	items, err := resource.List(context.Background(), "root:alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Metadata.Name != "a" {
		t.Fatalf("items = %+v", items)
	}
}

func TestCreatePostsTheObjectToItsNamespace(t *testing.T) {
	store, _ := newServer(t, []request{
		{Method: "POST", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 201, map[string]any{"metadata": map[string]any{"name": "run-1"}})
	})
	resource := Of[runObject](store, denoRuns)
	obj := &runObject{}
	obj.Metadata.Name = "run-1"
	obj.Metadata.Namespace = "default"
	if err := resource.Create(context.Background(), "root:alice", obj); err != nil {
		t.Fatal(err)
	}
}

func TestPatchStatusUsesTheSubresource(t *testing.T) {
	store, seen := newServer(t, []request{
		{Method: "PATCH", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns/run-1/status"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{})
	})
	resource := Of[runObject](store, denoRuns)
	patch, err := statuspatch.Merge(map[string]any{"phase": "Succeeded"})
	if err != nil {
		t.Fatal(err)
	}
	if err := resource.PatchStatus(context.Background(), ref.New("root:alice", "default", "run-1"), patch); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains((*seen)[0].Body, `"status"`) || !strings.Contains((*seen)[0].Body, "Succeeded") {
		t.Fatalf("body = %s", (*seen)[0].Body)
	}
	if !strings.Contains((*seen)[0].ContentType, "merge-patch") {
		t.Fatalf("content type = %q", (*seen)[0].ContentType)
	}
}

func TestRemoveFinalizerTestsTheCachedList(t *testing.T) {
	store, seen := newServer(t, []request{
		{Method: "GET", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns/run-1"},
		{Method: "PATCH", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns/run-1"},
	}, func(w http.ResponseWriter, _ *http.Request, index int) {
		if index == 0 {
			writeJSON(w, 200, map[string]any{
				"metadata": map[string]any{"name": "run-1", "finalizers": []string{"a", "b"}},
			})
			return
		}
		writeJSON(w, 200, map[string]any{})
	})
	resource := Of[runObject](store, denoRuns)
	if err := resource.RemoveFinalizer(context.Background(), ref.New("root:alice", "default", "run-1"), "a"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains((*seen)[1].Body, `"op":"test"`) {
		t.Fatalf("patch = %s", (*seen)[1].Body)
	}
}

func TestFinalizersAndKnownRemoval(t *testing.T) {
	store, _ := newServer(t, []request{
		{Method: "GET", Path: "/clusters/root:alice/apis/deno.computer/v1alpha1/namespaces/default/denoruns/run-1"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{
			"metadata": map[string]any{"name": "run-1", "finalizers": []string{"a"}},
		})
	})
	resource := Of[runObject](store, denoRuns)
	finalizers, err := resource.Finalizers(context.Background(), ref.New("root:alice", "default", "run-1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(finalizers) != 1 || finalizers[0] != "a" {
		t.Fatalf("finalizers = %v", finalizers)
	}
}

func TestNotFoundIsRecognised(t *testing.T) {
	store, _ := newServer(t, nil, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 404, map[string]any{"kind": "Status", "reason": "NotFound", "code": 404})
	})
	resource := Of[runObject](store, denoRuns)
	_, err := resource.Get(context.Background(), ref.New("root:alice", "default", "missing"))
	if err == nil || !IsNotFound(err) {
		t.Fatalf("err = %v", err)
	}
	if err := resource.Delete(context.Background(), ref.New("root:alice", "default", "missing")); err != nil {
		t.Fatalf("a delete of a missing object is not an error: %v", err)
	}
}

func TestMintServiceAccountToken(t *testing.T) {
	store, seen := newServer(t, []request{
		{Method: "POST", Path: "/clusters/root:alice/api/v1/namespaces/default/serviceaccounts/reader/token"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 201, map[string]any{"status": map[string]any{"token": "jwt-token"}})
	})
	token, err := store.MintServiceAccountToken(context.Background(), "root:alice", "default", "reader", 3600*1e9)
	if err != nil {
		t.Fatal(err)
	}
	if token != "jwt-token" {
		t.Fatalf("token = %q", token)
	}
	if !strings.Contains((*seen)[0].Body, "TokenRequest") {
		t.Fatalf("body = %s", (*seen)[0].Body)
	}
}

func TestClusterPathReadsTheAnnotation(t *testing.T) {
	store, _ := newServer(t, []request{
		{Method: "GET", Path: "/clusters/2j35/apis/core.kcp.io/v1alpha1/logicalclusters/cluster"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{
			"metadata": map[string]any{"annotations": map[string]any{"kcp.io/path": "root:alice"}},
		})
	})
	path, err := store.ClusterPath(context.Background(), "2j35")
	if err != nil {
		t.Fatal(err)
	}
	if path != "root:alice" {
		t.Fatalf("path = %q", path)
	}
}

func TestPathCacheResolvesOnce(t *testing.T) {
	store, seen := newServer(t, []request{
		{Method: "GET", Path: "/clusters/2j35/apis/core.kcp.io/v1alpha1/logicalclusters/cluster"},
	}, func(w http.ResponseWriter, _ *http.Request, _ int) {
		writeJSON(w, 200, map[string]any{
			"metadata": map[string]any{"annotations": map[string]any{"kcp.io/path": "root:alice"}},
		})
	})
	cache := NewPathCache(store)
	if path := cache.Lookup(context.Background(), "2j35"); path != "root:alice" {
		t.Fatalf("path = %q", path)
	}
	if path := cache.Lookup(context.Background(), "2j35"); path != "root:alice" {
		t.Fatalf("cached path = %q", path)
	}
	if len(*seen) != 1 {
		t.Fatalf("requests = %d, want the resolution to be cached", len(*seen))
	}
	cache.Forget("2j35")
	if cache.Len() != 0 {
		t.Fatal("forget must drop the entry")
	}
}

func TestNewRequiresAHost(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("a host is required")
	}
}
