package kcpstore

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

var liveProbes = schema.GroupVersionResource{Group: livekcp.Group, Version: livekcp.Version, Resource: livekcp.Resource}

type liveProbe struct {
	APIVersion string `json:"apiVersion"`

	Kind string `json:"kind"`

	Metadata struct {
		Name string `json:"name"`

		Namespace string `json:"namespace"`

		UID string `json:"uid"`

		ResourceVersion string `json:"resourceVersion"`

		Finalizers []string `json:"finalizers"`

		Generation int64 `json:"generation"`
	} `json:"metadata"`

	Spec struct {
		Steps int32 `json:"steps"`
	} `json:"spec"`

	Status struct {
		Phase string `json:"phase,omitempty"`

		Observed int32 `json:"observed,omitempty"`
	} `json:"status"`
}

func TestLiveKcpstoreAgainstRealKCP(t *testing.T) {
	livekcp.Require(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cluster, err := livekcp.Start(ctx)
	if err != nil {
		t.Fatalf("start kcp: %v", err)
	}
	defer cluster.Stop()

	store, err := New(Options{Host: cluster.Server, RestConfig: cluster.Config})
	if err != nil {
		t.Fatal(err)
	}
	probes := Of[liveProbe](store, liveProbes)
	target := ref.New(cluster.Consumer, "default", "alpha")

	err = probes.Create(ctx, cluster.Consumer, &liveProbe{
		APIVersion: livekcp.Group + "/" + livekcp.Version,
		Kind:       livekcp.Kind,
		Metadata: struct {
			Name string `json:"name"`

			Namespace string `json:"namespace"`

			UID string `json:"uid"`

			ResourceVersion string `json:"resourceVersion"`

			Finalizers []string `json:"finalizers"`

			Generation int64 `json:"generation"`
		}{Name: "alpha", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	created, err := probes.Get(ctx, target)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if created.Metadata.UID == "" || created.Metadata.ResourceVersion == "" {
		t.Fatalf("the server must stamp identity and a resource version: %+v", created.Metadata)
	}

	listed, err := probes.List(ctx, cluster.Consumer)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 || listed[0].Metadata.Name != "alpha" {
		t.Fatalf("listed = %+v", listed)
	}

	patch, err := StatusPatch(map[string]any{"phase": "Running", "observed": 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := probes.PatchStatus(ctx, target.WithResourceVersion(created.Metadata.ResourceVersion), patch); err != nil {
		t.Fatalf("patch status: %v", err)
	}
	patched, err := probes.Get(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if patched.Status.Phase != "Running" || patched.Status.Observed != 2 {
		t.Fatalf("status = %+v", patched.Status)
	}
	if patched.Spec.Steps != created.Spec.Steps {
		t.Fatal("a status patch must not touch the spec")
	}

	if err := probes.AddFinalizer(ctx, target, []string{"live.example.computer/held"}); err != nil {
		t.Fatalf("add finalizer: %v", err)
	}
	finalizers, err := probes.Finalizers(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(finalizers) != 1 || finalizers[0] != "live.example.computer/held" {
		t.Fatalf("finalizers = %v", finalizers)
	}
	if err := probes.RemoveFinalizer(ctx, target, "live.example.computer/held"); err != nil {
		t.Fatalf("remove finalizer: %v", err)
	}
	remaining, err := probes.Finalizers(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Fatalf("finalizers = %v, want none", remaining)
	}

	id, err := store.ClusterPath(ctx, cluster.Consumer)
	if err != nil {
		t.Fatalf("cluster path: %v", err)
	}
	if id != cluster.Consumer {
		t.Fatalf("path = %q, want %q", id, cluster.Consumer)
	}

	if err := probes.Delete(ctx, target); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := probes.Get(ctx, target); !IsNotFound(err) {
		t.Fatalf("get after delete = %v, want not found", err)
	}
}
