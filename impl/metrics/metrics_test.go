package metrics

import (
	"bytes"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRenderIncludesEveryFamily(t *testing.T) {
	registry := New("kcp")
	counter := registry.Counter("reconciles_total", "reconciles started")
	counter.Inc()
	counter.Add(2)
	gauge := registry.Gauge("queue_depth", "work keys waiting")
	gauge.Set(7)
	summary := registry.Summary("reconcile_seconds", "time inside a reconcile")
	summary.Observe(1.5)

	var buffer bytes.Buffer
	registry.Render(&buffer)
	body := buffer.String()
	for _, want := range []string{
		"# TYPE kcp_reconciles_total counter",
		"kcp_reconciles_total 3",
		"# TYPE kcp_queue_depth gauge",
		"kcp_queue_depth 7",
		"# TYPE kcp_reconcile_seconds summary",
		"kcp_reconcile_seconds_sum 1.5",
		"kcp_reconcile_seconds_count 1",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered metrics must contain %q, got:\n%s", want, body)
		}
	}
}

func TestRegisteringANameTwiceReturnsTheSameCollector(t *testing.T) {
	registry := New("kcp")
	first := registry.Counter("reconciles_total", "reconciles started")
	second := registry.Counter("reconciles_total", "reconciles started")
	first.Inc()
	if second != first {
		t.Fatal("a repeated name must return the collector that is already registered")
	}
	if got := testutil.ToFloat64(second); got != 1 {
		t.Fatalf("counter = %v, want 1", got)
	}
	var buffer bytes.Buffer
	registry.Render(&buffer)
	if strings.Count(buffer.String(), "# TYPE kcp_reconciles_total") != 1 {
		t.Fatalf("a repeated name must be described once:\n%s", buffer.String())
	}
}

func TestRegisteringANameWithADifferentHelpPanics(t *testing.T) {
	registry := New("kcp")
	registry.Counter("reconciles_total", "reconciles started")
	defer func() {
		if recover() == nil {
			t.Fatal("the help string is part of a metric's identity; a second one must not be swallowed")
		}
	}()
	registry.Counter("reconciles_total", "a different help")
}

func TestName(t *testing.T) {
	if got := New("kcp").Name("queue_depth"); got != "kcp_queue_depth" {
		t.Fatalf("name = %q", got)
	}
	if got := New("").Name("queue_depth"); got != "queue_depth" {
		t.Fatalf("name = %q", got)
	}
}

func TestListenServesMetrics(t *testing.T) {
	registry := New("kcp")
	registry.Counter("reconciles_total", "reconciles started").Inc()
	server, err := registry.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if server.Address() == "" {
		t.Fatal("the server must report its bound address")
	}
}

func TestTheRegistryIsAStandardRegisterer(t *testing.T) {
	registry := New("kcp")
	var _ prometheus.Registerer = registry.Registry()
	if registry.Registry() == nil {
		t.Fatal("the underlying registry must be reachable for anything this wrapper does not cover")
	}
}
