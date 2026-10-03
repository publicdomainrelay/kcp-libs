package cache

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/publicdomainrelay/kcp-libs/common/deno"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type fakeIndexer struct {
	objects map[string]any
}

func (f *fakeIndexer) GetByKey(key string) (any, bool, error) {
	obj, ok := f.objects[key]
	return obj, ok, nil
}

func (f *fakeIndexer) ByIndex(indexName, indexedValue string) ([]any, error) {
	var out []any
	for key, obj := range f.objects {
		values, err := IndexersFor(deno.PolicyWorkflowPodLabel, deno.JobRunLabel, "policyWorkflowPod")[indexName](obj)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if value == indexedValue {
				out = append(out, obj)
				break
			}
		}
		_ = key
	}
	return out, nil
}

func (f *fakeIndexer) List() []any {
	out := make([]any, 0, len(f.objects))
	for _, obj := range f.objects {
		out = append(out, obj)
	}
	return out
}

type runStatus struct {
	Phase string `json:"phase"`
}

type runObject struct {
	Metadata struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Labels    map[string]string `json:"labels"`
	} `json:"metadata"`
	Status runStatus `json:"status"`
}

func object(name, namespace, cluster, parent, phase string) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "deno.computer/v1alpha1",
		"kind":       "PolicyWorkflowRun",
		"metadata": map[string]any{
			"name":        name,
			"namespace":   namespace,
			"annotations": map[string]any{"kcp.io/cluster": cluster},
			"labels":      map[string]any{deno.PolicyWorkflowPodLabel: parent},
		},
		"status": map[string]any{"phase": phase},
	}
	return &unstructured.Unstructured{Object: obj}
}

func TestSetGetListAndIndex(t *testing.T) {
	set := NewSet()
	indexer := &fakeIndexer{objects: map[string]any{}}
	first := object("run-a", "default", "root:alice", "pod", "Running")
	second := object("run-b", "default", "root:alice", "other", "Pending")
	indexer.objects[ref.New("root:alice", "default", "run-a").Key()] = first
	indexer.objects[ref.New("root:alice", "default", "run-b").Key()] = second
	set.Add("policyworkflowrun", indexer)

	found, ok := set.Get("policyworkflowrun", ref.New("root:alice", "default", "run-a"))
	if !ok || found != first {
		t.Fatalf("get = (%v, %v)", found, ok)
	}
	if len(set.List("policyworkflowrun")) != 2 {
		t.Fatalf("list = %v", set.List("policyworkflowrun"))
	}
	byParent := set.ByIndex("policyworkflowrun", ByClusterParent, ref.Key("root:alice", "default", "pod"))
	if len(byParent) != 1 {
		t.Fatalf("by parent = %v", byParent)
	}
	if names := set.Names("policyworkflowrun", ByClusterParent, ref.Key("root:alice", "default", "other")); len(names) != 1 || names[0] != "run-b" {
		t.Fatalf("names = %v", names)
	}
}

func TestDecodeIntoConsumerTypes(t *testing.T) {
	obj := object("run-a", "default", "root:alice", "pod", "Running")
	decoded, err := Decode[runObject](obj)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Metadata.Name != "run-a" || decoded.Status.Phase != "Running" {
		t.Fatalf("decoded = %+v", decoded)
	}
	if decoded.Metadata.Labels[deno.PolicyWorkflowPodLabel] != "pod" {
		t.Fatalf("labels = %v", decoded.Metadata.Labels)
	}
	missing, err := Decode[runObject](nil)
	if err != nil || missing != nil {
		t.Fatalf("decoding nothing = (%v, %v)", missing, err)
	}
}

func TestObjectHelpers(t *testing.T) {
	obj := object("run-a", "default", "root:alice", "pod", "Running")
	if ClusterOf(obj) != "root:alice" {
		t.Fatalf("cluster = %q", ClusterOf(obj))
	}
	target, ok := RefOf(obj)
	if !ok || target.Name != "run-a" || target.Namespace != "default" {
		t.Fatalf("ref = (%+v, %v)", target, ok)
	}
	if PhaseOf(obj) != "Running" {
		t.Fatalf("phase = %q", PhaseOf(obj))
	}
	if NameOf(obj) != "run-a" {
		t.Fatalf("name = %q", NameOf(obj))
	}
	plain := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "x"}}}
	if _, ok := RefOf(plain); ok {
		t.Fatal("an object without the cluster annotation has no ref")
	}
}

func TestIndexersForClusterNameAndJob(t *testing.T) {
	indexers := IndexersFor(deno.PolicyWorkflowPodLabel, deno.JobRunLabel, "policyWorkflowPod")
	obj := object("run-a", "default", "root:alice", "pod", "Running")
	obj.SetLabels(map[string]string{deno.JobRunLabel: "job-1"})

	values, err := indexers[ByClusterName](obj)
	if err != nil || len(values) != 1 || values[0] != ref.Key("root:alice", "default", "run-a") {
		t.Fatalf("by-cluster-name = (%v, %v)", values, err)
	}
	values, err = indexers[ByClusterJob](obj)
	if err != nil || len(values) != 1 || values[0] != ref.Key("root:alice", "default", "job-1") {
		t.Fatalf("by-cluster-job = (%v, %v)", values, err)
	}
	values, err = indexers[ByCluster](obj)
	if err != nil || len(values) != 1 || values[0] != "root:alice" {
		t.Fatalf("by-cluster = (%v, %v)", values, err)
	}
}
