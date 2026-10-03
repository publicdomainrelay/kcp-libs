package metrics

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderIncludesEveryFamily(t *testing.T) {
	registry := New("kcp")
	counter := registry.Counter("reconciles_total", "reconciles started")
	counter.Inc()
	counter.Add(2)
	gauge := registry.Gauge("queue_depth", "work keys waiting")
	gauge.SetInt(7)
	summary := registry.Summary("reconcile_seconds", "time inside a reconcile")
	summary.Observe(1500 * time.Millisecond)

	var buffer bytes.Buffer
	registry.Render(&buffer)
	body := buffer.String()
	for _, want := range []string{
		"# TYPE kcp_reconciles_total counter",
		"kcp_reconciles_total 3",
		"# TYPE kcp_queue_depth gauge",
		"kcp_queue_depth 7",
		"# TYPE kcp_reconcile_seconds summary",
		"kcp_reconcile_seconds_sum 1.500000",
		"kcp_reconcile_seconds_count 1",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered metrics must contain %q, got:\n%s", want, body)
		}
	}
}

func TestRegistryIsIdempotentPerName(t *testing.T) {
	registry := New("kcp")
	first := registry.Counter("reconciles_total", "reconciles started")
	second := registry.Counter("reconciles_total", "a different help")
	first.Inc()
	if second.Value() != 1 {
		t.Fatalf("a repeated name must return the same counter, got %d", second.Value())
	}
	var buffer bytes.Buffer
	registry.Render(&buffer)
	if strings.Count(buffer.String(), "kcp_reconciles_total 1") != 1 {
		t.Fatalf("a repeated name must render once:\n%s", buffer.String())
	}
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
