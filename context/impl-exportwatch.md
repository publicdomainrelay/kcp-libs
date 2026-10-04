# Context: impl-exportwatch

Repository: `kcp-libs`

This context exists so callers can wait until a provider's exports are actually reachable before proceeding: it isolates the endpoint-slice discovery, projection and readiness logic from any workload runtime, gives a single place where the provider workspace is appended to the API host path, and exposes both a blocking wait (Await) and a watch-driven wait (WatchEndpointSlices) over the same Options.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: Await
  signature: func Await(ctx context.Context, opts Options) (Endpoints, error)
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: Client
  signature: func Client(opts Options) (dynamic.Interface, error)
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: Discover
  signature: func Discover(ctx context.Context, opts Options) (Endpoints, error)
- file: impl/exportwatch/exportwatch.go
  kind: type_alias
  name: Endpoints
  signature: type Endpoints map[string][]string
- file: impl/exportwatch/exportwatch.go
  kind: method
  name: Endpoints.Ready
  signature: func (e Endpoints) Ready(exports ...string) bool
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: FromList
  signature: func FromList(list *unstructured.UnstructuredList) Endpoints
- file: impl/exportwatch/exportwatch.go
  kind: struct
  name: Options
  signature: type Options struct
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: Paths
  signature: func Paths(endpoints Endpoints, export string) []string
- file: impl/exportwatch/exportwatch.go
  kind: function
  name: WatchEndpointSlices
  signature: func WatchEndpointSlices(ctx context.Context, opts Options, wait time.Duration)
    error
requirements:
- codeRefs:
  - function:0e22ac2739d17dd3caab2abd75c390e0
  - function:48d584db657cf9d26e0c376b34aa6773
  - struct:99a8a7585a6b5d30375c471a8fb38c71
  id: r.await-returns-discovered-endpoints
  level: MUST
  text: Await takes the same options as Discover and returns the discovered Endpoints
    once they are ready, so a blocking wait and a one-shot discovery share one configuration.
- codeRefs:
  - function:efb61af66b10e227581818288975a28e
  id: r.client-requires-config
  level: MUST
  text: 'Client refuses a nil rest config with the error "exportwatch: a rest config
    is required", so a client is never built without one.'
- codeRefs:
  - function:efb61af66b10e227581818288975a28e
  - struct:99a8a7585a6b5d30375c471a8fb38c71
  id: r.client-scopes-host-to-workspace
  level: MUST
  text: Client applies the client limit tuning to the config and rewrites the host
    to the options host plus the API path prefix plus the provider workspace, so every
    request is scoped to that workspace.
- codeRefs:
  - function:0e22ac2739d17dd3caab2abd75c390e0
  - function:4855b61ac47e7bbc0bb211b8661e9d59
  id: r.discover-lists-endpoint-slices
  level: MUST
  text: 'Discover builds a client from the options, lists the endpoint slice resource,
    and wraps a list failure as "exportwatch: list the endpoint slices: %w"; on success
    it returns FromList of the result.'
- codeRefs:
  - file:impl/exportwatch/exportwatch.go
  - type_alias:94252f87faff3176628078b9500af709
  id: r.endpoints-map-shape
  level: MUST
  text: Endpoint addresses are held as a map from export name to a slice of URL strings,
    so one export can carry several addresses.
- codeRefs:
  - function:4855b61ac47e7bbc0bb211b8661e9d59
  id: r.fromlist-reads-export-and-urls
  level: MUST
  text: FromList reads spec.export.name from each item and the url of each entry in
    status.endpoints, skipping items with no export name and entries with no url,
    and appends the urls under that export name.
- codeRefs:
  - function:efb61af66b10e227581818288975a28e
  - struct:99a8a7585a6b5d30375c471a8fb38c71
  id: r.options-carries-host-resolution
  level: MAY
  text: Options exposes an unexported host method that yields the API host, so the
    host derivation stays with the options that describe the provider workspace.
- codeRefs:
  - function:96edd4eb5a20bcded694ab307fa5f365
  - type_alias:94252f87faff3176628078b9500af709
  id: r.paths-renders-one-export
  level: MUST
  text: Paths takes the discovered Endpoints and a single export name and returns
    that export's addresses as a string slice.
- codeRefs:
  - method:631983bf08fd7d3133061275ab1f910d
  id: r.ready-requires-every-export
  level: MUST
  text: Endpoints.Ready reports false when any named export has zero addresses and
    true only when every named export has at least one.
- codeRefs:
  - function:b185d3ac6de1d5bfcfd671f7afeca0c2
  id: r.watch-bounds-itself
  level: MUST
  text: WatchEndpointSlices derives a context bounded by the wait duration, stops
    the watch on return, and reports success (nil) when that context ends or the result
    channel closes.
- codeRefs:
  - function:0e22ac2739d17dd3caab2abd75c390e0
  - function:b185d3ac6de1d5bfcfd671f7afeca0c2
  - method:631983bf08fd7d3133061275ab1f910d
  id: r.watch-signals-readiness
  level: MUST
  text: On an Added, Modified or Deleted event WatchEndpointSlices rediscovers the
    endpoints and returns ErrReady as soon as the options exports are ready; a failed
    rediscovery is ignored rather than fatal.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/exportwatch/exportwatch.go` file exportwatch.go (impl/exportwatch/exportwatch.go)
- `function:0e22ac2739d17dd3caab2abd75c390e0` function Discover (impl/exportwatch/exportwatch.go)
- `function:4855b61ac47e7bbc0bb211b8661e9d59` function FromList (impl/exportwatch/exportwatch.go)
- `function:48d584db657cf9d26e0c376b34aa6773` function Await (impl/exportwatch/exportwatch.go)
- `function:96edd4eb5a20bcded694ab307fa5f365` function Paths (impl/exportwatch/exportwatch.go)
- `function:b185d3ac6de1d5bfcfd671f7afeca0c2` function WatchEndpointSlices (impl/exportwatch/exportwatch.go)
- `function:efb61af66b10e227581818288975a28e` function Client (impl/exportwatch/exportwatch.go)
- `method:631983bf08fd7d3133061275ab1f910d` method Endpoints.Ready (impl/exportwatch/exportwatch.go)
- `struct:99a8a7585a6b5d30375c471a8fb38c71` struct Options (impl/exportwatch/exportwatch.go)
- `type_alias:94252f87faff3176628078b9500af709` type_alias Endpoints (impl/exportwatch/exportwatch.go)
<!-- SPECD_MANAGED_END -->
