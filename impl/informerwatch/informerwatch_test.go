package informerwatch

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

func gvr(resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "example.computer", Version: "v1alpha1", Resource: resource}
}

func config(t *testing.T) *rest.Config {
	t.Helper()
	return &rest.Config{Host: "https://kcp.invalid"}
}

func widget(cluster, namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.computer/v1alpha1",
		"kind":       "Widget",
		"metadata": map[string]any{
			"name":        name,
			"namespace":   namespace,
			"annotations": map[string]any{"kcp.io/cluster": cluster},
		},
	}}
}

func TestEnqueueOwnReadsTheClusterAnnotation(t *testing.T) {
	var got []ref.Ref
	var kinds []string
	enqueue := func(kind string, r ref.Ref) {
		kinds = append(kinds, kind)
		got = append(got, r)
	}
	enqueueOwn("widget", widget("root:alice", "default", "one"), enqueue)
	if len(got) != 1 || kinds[0] != "widget" {
		t.Fatalf("enqueued = %v %v", kinds, got)
	}
	if got[0].LogicalCluster != "root:alice" || got[0].Namespace != "default" || got[0].Name != "one" {
		t.Fatalf("ref = %+v", got[0])
	}
}

func TestEnqueueOwnSkipsAnObjectWithNoCluster(t *testing.T) {
	enqueued := 0
	bare := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "one"}}}
	enqueueOwn("widget", bare, func(string, ref.Ref) { enqueued++ })
	if enqueued != 0 {
		t.Fatal("an object without the cluster annotation has no place in the queue")
	}
}

func TestToK8sIndexersCarriesEveryIndex(t *testing.T) {
	indexers := cache.IndexersFor("example.computer/group", "example.computer/job", "parent")
	converted := toK8sIndexers(indexers)
	if len(converted) != len(indexers) {
		t.Fatalf("converted = %d, want %d", len(converted), len(indexers))
	}
	byCluster, ok := converted[cache.ByCluster]
	if !ok {
		t.Fatalf("by-cluster was dropped: %v", converted)
	}
	values, err := byCluster(widget("root:alice", "default", "one"))
	if err != nil || len(values) != 1 || values[0] != "root:alice" {
		t.Fatalf("by-cluster = (%v, %v)", values, err)
	}
	byName, ok := converted[cache.ByClusterName]
	if !ok {
		t.Fatalf("by-cluster-name was dropped: %v", converted)
	}
	values, err = byName(widget("root:alice", "default", "one"))
	if err != nil || len(values) != 1 || values[0] != "root:alice/default/one" {
		t.Fatalf("by-cluster-name = (%v, %v)", values, err)
	}
}

func TestRunRefusesAnIncompleteConfiguration(t *testing.T) {
	if err := Run(t.Context(), Options{}); err == nil {
		t.Fatal("a rest config is required")
	}
	if err := Run(t.Context(), Options{Config: config(t)}); err == nil {
		t.Fatal("an enqueue function is required")
	}
}

func TestAWatchSourceMustNameABaseAndItsResources(t *testing.T) {
	config := &rest.Config{Host: "https://kcp.invalid"}
	enqueue := func(string, ref.Ref) {}
	cases := map[string]Options{
		"no base":     {Config: config, Enqueue: enqueue, Sources: []Source{{Resources: []Resource{{Kind: "widget", GVR: gvr("widgets")}}}}},
		"no resource": {Config: config, Enqueue: enqueue, Sources: []Source{{Base: "https://kcp.example/widgets"}}},
	}
	for name, opts := range cases {
		if err := Run(t.Context(), opts); err == nil {
			t.Fatalf("%s must be refused", name)
		}
	}
}

func TestAKindWatchedTwiceIsRefused(t *testing.T) {
	source := func(base string) Source {
		return Source{Base: base, Resources: []Resource{{Kind: "widget", GVR: gvr("widgets")}}}
	}
	opts := Options{
		Config:  config(t),
		Enqueue: func(string, ref.Ref) {},
		Sources: []Source{source("https://kcp.example/a"), source("https://kcp.example/b")},
	}
	err := Run(t.Context(), opts)
	if err == nil {
		t.Fatal("a kind watched from two sources would have one of its indexes replaced by the other")
	}
	if !strings.Contains(err.Error(), "watched twice") {
		t.Fatalf("err = %v", err)
	}
}
