package cache

import (
	"encoding/json"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/publicdomainrelay/kcp-libs/common/kcp"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

const (
	ByCluster = "by-cluster"

	ByClusterName = "by-cluster-name"

	ByClusterParent = "by-cluster-parent"

	ByClusterJob = "by-cluster-job"

	ByClusterTriggerPod = "by-cluster-trigger-pod"
)

type Indexer interface {
	GetByKey(key string) (item any, exists bool, err error)

	ByIndex(indexName, indexedValue string) ([]any, error)

	List() []any
}

type IndexFunc func(obj any) ([]string, error)

type Indexers map[string]IndexFunc

type Set struct {
	mu sync.RWMutex

	indexers map[string][]Indexer
}

func NewSet() *Set {
	return &Set{indexers: map[string][]Indexer{}}
}

func (s *Set) Add(kind string, indexer Indexer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.indexers[kind] = append(s.indexers[kind], indexer)
}

func (s *Set) of(kind string) []Indexer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Indexer(nil), s.indexers[kind]...)
}

func (s *Set) Registered(kind string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.indexers[kind]) > 0
}

func (s *Set) Get(kind string, r ref.Ref) (any, bool) {
	key := r.Key()
	for _, indexer := range s.of(kind) {
		obj, exists, err := indexer.GetByKey(key)
		if err == nil && exists {
			return obj, true
		}
	}
	return nil, false
}

func (s *Set) List(kind string) []any {
	var out []any
	for _, indexer := range s.of(kind) {
		out = append(out, indexer.List()...)
	}
	return out
}

func (s *Set) ByIndex(kind, index, value string) []any {
	var out []any
	for _, indexer := range s.of(kind) {
		objs, err := indexer.ByIndex(index, value)
		if err != nil {
			continue
		}
		out = append(out, objs...)
	}
	return out
}

func (s *Set) Names(kind, index, value string) []string {
	var out []string
	for _, obj := range s.ByIndex(kind, index, value) {
		if name := NameOf(obj); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func Decode[T any](obj any) (*T, error) {
	if obj == nil {
		return nil, nil
	}
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("cache: encode object: %w", err)
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("cache: decode object: %w", err)
	}
	return &out, nil
}

func ClusterOf(obj any) string {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ""
	}
	return u.GetAnnotations()[kcp.ClusterAnnotation]
}

func RefOf(obj any) (ref.Ref, bool) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ref.Ref{}, false
	}
	cluster := u.GetAnnotations()[kcp.ClusterAnnotation]
	if cluster == "" {
		return ref.Ref{}, false
	}
	return ref.New(cluster, u.GetNamespace(), u.GetName()), true
}

func NameOf(obj any) string {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ""
	}
	return u.GetName()
}

func PhaseOf(obj any) string {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ""
	}
	phase, _, _ := unstructured.NestedString(u.Object, "status", "phase")
	return phase
}

func NestedString(obj any, fields ...string) string {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ""
	}
	value, _, _ := unstructured.NestedString(u.Object, fields...)
	return value
}

func IndexersFor(parentLabel, jobLabel, triggerPodField string) Indexers {
	return Indexers{
		ByCluster: func(obj any) ([]string, error) {
			cluster := ClusterOf(obj)
			if cluster == "" {
				return nil, nil
			}
			return []string{cluster}, nil
		},
		ByClusterName: func(obj any) ([]string, error) {
			r, ok := RefOf(obj)
			if !ok {
				return nil, nil
			}
			return []string{r.Key()}, nil
		},
		ByClusterParent: func(obj any) ([]string, error) {
			r, ok := RefOf(obj)
			if !ok {
				return nil, nil
			}
			u := obj.(*unstructured.Unstructured)
			parent := u.GetLabels()[parentLabel]
			if parent == "" {
				return nil, nil
			}
			return []string{ref.Key(r.LogicalCluster, r.Namespace, parent)}, nil
		},
		ByClusterJob: func(obj any) ([]string, error) {
			r, ok := RefOf(obj)
			if !ok {
				return nil, nil
			}
			u := obj.(*unstructured.Unstructured)
			job := u.GetLabels()[jobLabel]
			if job == "" {
				return nil, nil
			}
			return []string{ref.Key(r.LogicalCluster, r.Namespace, job)}, nil
		},
		ByClusterTriggerPod: func(obj any) ([]string, error) {
			r, ok := RefOf(obj)
			if !ok {
				return nil, nil
			}
			parent := NestedString(obj, triggerPodField)
			if parent == "" {
				return nil, nil
			}
			return []string{ref.Key(r.LogicalCluster, r.Namespace, parent)}, nil
		},
	}
}
