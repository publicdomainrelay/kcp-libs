package kcp

import "testing"

func TestServiceLabels(t *testing.T) {
	cases := map[string]string{
		"root":             "",
		"root:alice":       "alice",
		"root:acme:prod":   "prod.acme",
		"root:acme:prod:x": "x.prod.acme",
	}
	for in, want := range cases {
		if got := ServiceLabels(in); got != want {
			t.Fatalf("ServiceLabels(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestServiceFQDN(t *testing.T) {
	cases := []struct {
		name      string
		namespace string
		cluster   string
		want      string
	}{
		{"pds", "default", "root:alice", "pds.default.alice.svc.kcp.local"},
		{"pds", "", "root:alice", "pds.default.alice.svc.kcp.local"},
		{"pds", "default", "root", "pds.default.svc.kcp.local"},
	}
	for _, tc := range cases {
		if got := ServiceFQDN(tc.name, tc.namespace, tc.cluster, DefaultServiceDomain); got != tc.want {
			t.Fatalf("ServiceFQDN = %q, want %q", got, tc.want)
		}
	}
}

func TestClusterFromLabelsInvertsServiceLabels(t *testing.T) {
	for _, cluster := range []string{"root:alice", "root:acme:prod", "root"} {
		if got := ClusterFromLabels(ServiceLabels(cluster)); got != cluster {
			t.Fatalf("round trip of %q = %q", cluster, got)
		}
	}
}

func TestSplitServiceFQDN(t *testing.T) {
	name, namespace, cluster, ok := SplitServiceFQDN("pds.default.alice.svc.kcp.local", DefaultServiceDomain)
	if !ok || name != "pds" || namespace != "default" || cluster != "root:alice" {
		t.Fatalf("split = (%q, %q, %q, %v)", name, namespace, cluster, ok)
	}
	if _, _, _, ok := SplitServiceFQDN("pds.default.other.example", DefaultServiceDomain); ok {
		t.Fatal("a host outside the suffix must not split")
	}
	if _, _, _, ok := SplitServiceFQDN("pds.svc.kcp.local", DefaultServiceDomain); ok {
		t.Fatal("a host without a namespace must not split")
	}
}
