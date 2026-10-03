package metrics

import (
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const contentType = "text/plain; version=0.0.4"

type Counter struct {
	value atomic.Uint64
}

func (c *Counter) Add(delta uint64) {
	c.value.Add(delta)
}

func (c *Counter) Inc() {
	c.value.Add(1)
}

func (c *Counter) Value() uint64 {
	return c.value.Load()
}

type Gauge struct {
	value atomic.Uint64
}

func (g *Gauge) Set(value float64) {
	g.value.Store(math.Float64bits(value))
}

func (g *Gauge) SetInt(value int64) {
	g.Set(float64(value))
}

func (g *Gauge) Add(delta float64) {
	for {
		old := g.value.Load()
		next := math.Float64bits(math.Float64frombits(old) + delta)
		if g.value.CompareAndSwap(old, next) {
			return
		}
	}
}

func (g *Gauge) Raise(value float64) {
	for {
		old := g.value.Load()
		if math.Float64frombits(old) >= value {
			return
		}
		if g.value.CompareAndSwap(old, math.Float64bits(value)) {
			return
		}
	}
}

func (g *Gauge) Value() float64 {
	return math.Float64frombits(g.value.Load())
}

type Summary struct {
	sum atomic.Uint64

	count atomic.Uint64
}

func (s *Summary) Observe(d time.Duration) {
	s.ObserveSeconds(d.Seconds())
}

func (s *Summary) ObserveSeconds(seconds float64) {
	for {
		old := s.sum.Load()
		next := math.Float64bits(math.Float64frombits(old) + seconds)
		if s.sum.CompareAndSwap(old, next) {
			break
		}
	}
	s.count.Add(1)
}

func (s *Summary) Sum() float64 {
	return math.Float64frombits(s.sum.Load())
}

func (s *Summary) Count() uint64 {
	return s.count.Load()
}

type metric interface {
	render(w io.Writer, name, help string)
}

type entry struct {
	name string

	help string

	metric metric
}

type Registry struct {
	prefix string

	mu sync.Mutex

	entries map[string]entry

	order []string
}

func New(prefix string) *Registry {
	return &Registry{prefix: prefix, entries: map[string]entry{}}
}

func (r *Registry) Counter(name, help string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.entries[name]; ok {
		if counter, ok := existing.metric.(*Counter); ok {
			return counter
		}
		counter := &Counter{}
		r.entries[name] = entry{name: name, help: help, metric: counter}
		return counter
	}
	counter := &Counter{}
	r.entries[name] = entry{name: name, help: help, metric: counter}
	r.order = append(r.order, name)
	return counter
}

func (r *Registry) Gauge(name, help string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.entries[name]; ok {
		if gauge, ok := existing.metric.(*Gauge); ok {
			return gauge
		}
		gauge := &Gauge{}
		r.entries[name] = entry{name: name, help: help, metric: gauge}
		return gauge
	}
	gauge := &Gauge{}
	r.entries[name] = entry{name: name, help: help, metric: gauge}
	r.order = append(r.order, name)
	return gauge
}

func (r *Registry) Summary(name, help string) *Summary {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.entries[name]; ok {
		if summary, ok := existing.metric.(*Summary); ok {
			return summary
		}
		summary := &Summary{}
		r.entries[name] = entry{name: name, help: help, metric: summary}
		return summary
	}
	summary := &Summary{}
	r.entries[name] = entry{name: name, help: help, metric: summary}
	r.order = append(r.order, name)
	return summary
}

func (r *Registry) Name(name string) string {
	if r.prefix == "" {
		return name
	}
	return r.prefix + "_" + name
}

func (r *Registry) Render(w io.Writer) {
	r.mu.Lock()
	entries := make([]entry, 0, len(r.order))
	for _, name := range r.order {
		entries = append(entries, r.entries[name])
	}
	prefix := r.prefix
	r.mu.Unlock()
	sort.SliceStable(entries, func(a, b int) bool { return entries[a].name < entries[b].name })
	for _, item := range entries {
		name := item.name
		if prefix != "" {
			name = prefix + "_" + name
		}
		item.metric.render(w, name, item.help)
	}
}

func (c *Counter) render(w io.Writer, name, help string) {
	fmt.Fprintf(w, "# HELP %s %s\n", name, help)
	fmt.Fprintf(w, "# TYPE %s counter\n", name)
	fmt.Fprintf(w, "%s %d\n", name, c.Value())
}

func (g *Gauge) render(w io.Writer, name, help string) {
	fmt.Fprintf(w, "# HELP %s %s\n", name, help)
	fmt.Fprintf(w, "# TYPE %s gauge\n", name)
	fmt.Fprintf(w, "%s %.6g\n", name, g.Value())
}

func (s *Summary) render(w io.Writer, name, help string) {
	fmt.Fprintf(w, "# HELP %s %s\n", name, help)
	fmt.Fprintf(w, "# TYPE %s summary\n", name)
	fmt.Fprintf(w, "%s_sum %.6f\n", name, s.Sum())
	fmt.Fprintf(w, "%s_count %d\n", name, s.Count())
}

func (r *Registry) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		r.Render(w)
	})
	return mux
}

type Server struct {
	listener net.Listener

	server *http.Server
}

func (r *Registry) Listen(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("metrics: bind %s: %w", address, err)
	}
	server := &http.Server{Handler: r.Handler()}
	go func() {
		_ = server.Serve(listener)
	}()
	return &Server{listener: listener, server: server}, nil
}

func (s *Server) Address() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Close()
}
