# Context: impl-metrics

Repository: `kcp-libs`

This context exists so that every component in the repository can expose Prometheus metrics through one small, uniform surface instead of touching the prometheus client directly. It gives each component an isolated registry plus a naming prefix, so two components in one process cannot collide on a metric name, and it makes the common collectors idempotent by name so repeated wiring is safe. It also supplies the two export paths a process needs: Render for in-process encoding and Listen for an HTTP scrape endpoint, with Server.Address and Server.Close for lifecycle control and tests that bind port zero.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: impl/metrics/metrics.go
  kind: function
  name: New
  signature: func New(prefix string) *Registry
- file: impl/metrics/metrics.go
  kind: struct
  name: Registry
  signature: type Registry struct
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Counter
  signature: func (r *Registry) Counter(name, help string) prometheus.Counter
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Gauge
  signature: func (r *Registry) Gauge(name, help string) prometheus.Gauge
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.GaugeFunc
  signature: func (r *Registry) GaugeFunc(name, help string, value func() float64)
    prometheus.GaugeFunc
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Handler
  signature: func (r *Registry) Handler() http.Handler
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Listen
  signature: func (r *Registry) Listen(address string) (*Server, error)
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Name
  signature: func (r *Registry) Name(name string) string
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Registry
  signature: func (r *Registry) Registry() *prometheus.Registry
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Render
  signature: func (r *Registry) Render(w io.Writer) error
- file: impl/metrics/metrics.go
  kind: method
  name: Registry.Summary
  signature: func (r *Registry) Summary(name, help string) prometheus.Summary
- file: impl/metrics/metrics.go
  kind: struct
  name: Server
  signature: type Server struct
- file: impl/metrics/metrics.go
  kind: method
  name: Server.Address
  signature: func (s *Server) Address() string
- file: impl/metrics/metrics.go
  kind: method
  name: Server.Close
  signature: func (s *Server) Close() error
requirements:
- codeRefs:
  - method:22dbba2741b262a6e8f413f68c6881a9
  id: r.gaugefunc-exclusive
  level: MUST
  text: Registry.GaugeFunc reads its value at scrape time, is bound to one caller,
    and must therefore panic on any registration error instead of sharing a collector,
    so two gauge functions cannot take the same name.
- codeRefs:
  - method:85f974cd4e6426f92ba1d7e4a6f5c3e7
  id: r.http-handler
  level: MUST
  text: Registry.Handler returns a promhttp handler over this registry alone, so an
    HTTP scrape reports only the metrics this Registry owns.
- codeRefs:
  - function:3dd18f3afa474317bc20bf9d8ce521ed
  - method:398ced1f3e757d374616550142bd5518
  - struct:7917d94287fa848087ff53481834da21
  id: r.isolated-registry
  level: MUST
  text: New creates a fresh prometheus.Registry, so a Registry never inherits metrics
    registered elsewhere, and Registry.Registry exposes it to callers that need the
    underlying collector registry.
- codeRefs:
  - method:efe5348355d57e0d4f64052375fb2ca7
  - struct:a14e163798928aaf7cacdaa7f5b3e1cb
  id: r.listen-binds-and-serves
  level: MUST
  text: 'Registry.Listen binds the requested TCP address, returns the bind error wrapped
    as "metrics: bind %s: %w", and otherwise serves Registry.Handler on that listener
    from a background goroutine behind the returned Server.'
- codeRefs:
  - function:3dd18f3afa474317bc20bf9d8ce521ed
  - method:75c11a332140d7b3055ad5d064bf82cf
  id: r.prefix-names
  level: MUST
  text: Registry.Name returns the name unchanged when the prefix is empty and prefix+"_"+name
    otherwise, and every collector constructor must use it so all exported metric
    names carry the prefix.
- codeRefs:
  - method:b5bbba2315db1d50de247c1650bdff0c
  id: r.render-text-plain
  level: MUST
  text: 'Registry.Render gathers the registry and encodes every family to the writer
    in text/plain exposition format, wrapping a gather failure as "metrics: gather:
    %w" and an encode failure as "metrics: encode %s: %w" with the family name.'
- codeRefs:
  - method:5aa29e31405d7485061154e56775d604
  - method:dafe620a88e59c54431ba7b2aa5c720e
  id: r.server-lifecycle-nil-safe
  level: MUST
  text: Server.Address returns the bound listener address, or the empty string for
    a nil Server or one with no listener, and Server.Close closes the HTTP server
    and returns nil for a nil Server or one with no server, so both are safe on a
    value that was never started.
- codeRefs:
  - method:54ddf7ba8219fb9150866ee200bac2ce
  - method:7c6900816ec7f4f0413e76c1f65af404
  - method:b518b6626d0caa622f7f0fa1c0410bcf
  id: r.shared-collectors-idempotent
  level: MUST
  text: Registry.Counter, Registry.Gauge and Registry.Summary must return the already
    registered collector when the same prefixed name is requested twice, and must
    panic when registration fails for any other reason, including a name registered
    with a different help string or a different collector type.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/metrics/metrics.go` file metrics.go (impl/metrics/metrics.go)
- `file:impl/metrics/metrics_test.go` file metrics_test.go (impl/metrics/metrics_test.go)
- `function:3dd18f3afa474317bc20bf9d8ce521ed` function New (impl/metrics/metrics.go)
- `method:22dbba2741b262a6e8f413f68c6881a9` method Registry.GaugeFunc (impl/metrics/metrics.go)
- `method:398ced1f3e757d374616550142bd5518` method Registry.Registry (impl/metrics/metrics.go)
- `method:54ddf7ba8219fb9150866ee200bac2ce` method Registry.Summary (impl/metrics/metrics.go)
- `method:5aa29e31405d7485061154e56775d604` method Server.Address (impl/metrics/metrics.go)
- `method:75c11a332140d7b3055ad5d064bf82cf` method Registry.Name (impl/metrics/metrics.go)
- `method:7c6900816ec7f4f0413e76c1f65af404` method Registry.Gauge (impl/metrics/metrics.go)
- `method:85f974cd4e6426f92ba1d7e4a6f5c3e7` method Registry.Handler (impl/metrics/metrics.go)
- `method:b518b6626d0caa622f7f0fa1c0410bcf` method Registry.Counter (impl/metrics/metrics.go)
- `method:b5bbba2315db1d50de247c1650bdff0c` method Registry.Render (impl/metrics/metrics.go)
- `method:dafe620a88e59c54431ba7b2aa5c720e` method Server.Close (impl/metrics/metrics.go)
- `method:efe5348355d57e0d4f64052375fb2ca7` method Registry.Listen (impl/metrics/metrics.go)
- `struct:7917d94287fa848087ff53481834da21` struct Registry (impl/metrics/metrics.go)
- `struct:a14e163798928aaf7cacdaa7f5b3e1cb` struct Server (impl/metrics/metrics.go)
<!-- SPECD_MANAGED_END -->
