# Context: impl-kcpstore

Repository: `kcp-libs`

This context exists so that consumers (controllers, admission, DNS, servicenames) have one place that knows how to reach kcp: URL shape, per-cluster client reuse, rate limiting, JSON decoding, status and finalizer patching, token minting, and workspace path resolution. The path cache is the newest part and it is deliberately conservative — it exists because a workspace path that failed to resolve must not be frozen into a wrong name, so only an answer that cannot change by asking again is stored.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: impl/kcpstore/kcpstore.go
  kind: function
  name: IsNotFound
  signature: func IsNotFound(err error) bool
- file: impl/kcpstore/kcpstore.go
  kind: function
  name: New
  signature: func New(opts Options) (*Store, error)
- file: impl/kcpstore/paths.go
  kind: function
  name: NewPathCache
  signature: func NewPathCache(store *Store) *PathCache
- file: impl/kcpstore/kcpstore.go
  kind: function
  name: Of
  signature: func Of[T any](s *Store, gvr schema.GroupVersionResource) *Resource[T]
- file: impl/kcpstore/kcpstore.go
  kind: struct
  name: Options
  signature: type Options struct
- file: impl/kcpstore/paths.go
  kind: struct
  name: PathCache
  signature: type PathCache struct
- file: impl/kcpstore/paths.go
  kind: method
  name: PathCache.Forget
  signature: func (c *PathCache) Forget(logicalCluster string)
- file: impl/kcpstore/paths.go
  kind: method
  name: PathCache.Len
  signature: func (c *PathCache) Len() int
- file: impl/kcpstore/paths.go
  kind: method
  name: PathCache.Lookup
  signature: func (c *PathCache) Lookup(ctx context.Context, logicalCluster string)
    string
- file: impl/kcpstore/kcpstore.go
  kind: struct
  name: Resource
  signature: type Resource[T any] struct
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.AddFinalizer
  signature: func (r *Resource[T]) AddFinalizer(ctx context.Context, target ref.Ref,
    finalizers []string) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.Create
  signature: func (r *Resource[T]) Create(ctx context.Context, logicalCluster string,
    obj *T) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.Delete
  signature: func (r *Resource[T]) Delete(ctx context.Context, target ref.Ref) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.Finalizers
  signature: func (r *Resource[T]) Finalizers(ctx context.Context, target ref.Ref)
    ([]string, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.Get
  signature: func (r *Resource[T]) Get(ctx context.Context, target ref.Ref) (*T, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.GetRaw
  signature: func (r *Resource[T]) GetRaw(ctx context.Context, target ref.Ref) ([]byte,
    error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.List
  signature: func (r *Resource[T]) List(ctx context.Context, logicalCluster string)
    ([]T, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.ListIn
  signature: func (r *Resource[T]) ListIn(ctx context.Context, logicalCluster, namespace
    string) ([]T, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.PatchJSON
  signature: func (r *Resource[T]) PatchJSON(ctx context.Context, target ref.Ref,
    patch []byte) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.PatchStatus
  signature: func (r *Resource[T]) PatchStatus(ctx context.Context, target ref.Ref,
    patch []byte) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.RemoveFinalizer
  signature: func (r *Resource[T]) RemoveFinalizer(ctx context.Context, target ref.Ref,
    finalizer string) error
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Resource.RemoveKnownFinalizer
  signature: func (r *Resource[T]) RemoveKnownFinalizer(ctx context.Context, target
    ref.Ref, current []string, finalizer string) error
- file: impl/kcpstore/kcpstore.go
  kind: struct
  name: Store
  signature: type Store struct
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Store.ClientFor
  signature: func (s *Store) ClientFor(logicalCluster string, gv schema.GroupVersion)
    (rest.Interface, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Store.ClusterPath
  signature: func (s *Store) ClusterPath(ctx context.Context, logicalCluster string)
    (string, error)
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Store.Config
  signature: func (s *Store) Config() *rest.Config
- file: impl/kcpstore/kcpstore.go
  kind: method
  name: Store.MintServiceAccountToken
  signature: func (s *Store) MintServiceAccountToken(ctx context.Context, logicalCluster,
    namespace, name string, ttl time.Duration) (string, error)
requirements:
- codeRefs:
  - method:5763969969c0f4e887f1c499a7fccf73
  id: r.add-finalizer
  level: MUST
  text: Resource.AddFinalizer must build the addition patch with statuspatch.FinalizerAdd
    and apply it through PatchJSON.
- codeRefs:
  - struct:432c046ab7a1ce292d7fa0372790eed7
  - struct:bd87c7e07b92542841411a9474f66987
  id: r.cache-mutex
  level: MUST
  text: PathCache and Store must guard their maps with a mutex, and neither Lookup
    nor ClientFor may hold the lock across the network call.
- codeRefs:
  - method:c311b4200adb69d996ac127afb62ed56
  - struct:432c046ab7a1ce292d7fa0372790eed7
  id: r.client-per-cluster-gv
  level: MUST
  text: ClientFor must key clients on logicalCluster plus the group-version string,
    return a cached client when one exists, and otherwise build one from a copy of
    the store config with APIPath set to ref.APIPathPrefix + logicalCluster + "/apis",
    or + "/api" when the group is empty.
- codeRefs:
  - method:22f545280ed249562e9ca742d5a77c44
  id: r.cluster-path-annotation
  level: MUST
  text: Store.ClusterPath must read the logicalclusters/cluster object in core.kcp.io/v1alpha1
    and return its kcp.PathAnnotation value, failing with ErrNoPath when the annotation
    is empty and wrapping read and parse failures with the logical cluster named.
- codeRefs:
  - method:c00fe3e383047902bea23ff36069e91f
  id: r.config-accessor
  level: MUST
  text: Store.Config must return the store's tuned rest.Config.
- codeRefs:
  - function:41b4ccba5b43f3162c43c6569fe3adaf
  - struct:be2a24917d7a285dc53d560ba36c744a
  id: r.config-tuning
  level: MUST
  text: New must derive the store config by applying the QPS and Burst limits to the
    supplied RestConfig, or to an empty config when RestConfig is nil, then set Host
    from ref.BaseHost, JSON content type and accept types, and the supplied Transport
    when non-nil.
- codeRefs:
  - method:e54d2aabb8486cab3ffb3fccd9e528b4
  id: r.create-from-metadata
  level: MUST
  text: Resource.Create must marshal the object, read its namespace from the encoded
    metadata, POST it into that namespace, and on failure include the response body
    detail plus the object name in the wrapped error.
- codeRefs:
  - function:d7d93064d1945db238404c68788f971f
  - method:0adeb1384568f6dec86b61baf285da43
  id: r.delete-background
  level: MUST
  text: Resource.Delete must delete with propagationPolicy Background and treat an
    already-absent object as success, returning nil for IsNotFound errors.
- codeRefs:
  - file:impl/kcpstore/kcpstore.go
  id: r.error-wrapping-prefix
  level: SHOULD
  text: Errors leaving the package should carry the "kcpstore:" prefix and name the
    resource, object, and logical cluster involved so callers can attribute a failure
    without re-reading the request.
- codeRefs:
  - method:cfd282c2aae4ed3dd3a523783c18bc20
  id: r.finalizers-read
  level: MUST
  text: Resource.Finalizers must read the raw object and return the finalizers parsed
    by statuspatch.FinalizersOf.
- codeRefs:
  - function:96752734b0a49c97d1a14ee1329189da
  - struct:dafdd052ffecc9d0a327766098dc06f3
  id: r.generic-resource
  level: MUST
  text: Of must return a Resource[T] bound to a Store and a GroupVersionResource,
    and every Resource method must route through the client for the target's logical
    cluster.
- codeRefs:
  - method:28c264fe91c941c628570295cc91bcea
  id: r.get-decode
  level: MUST
  text: 'Resource.Get must fetch the raw object and unmarshal it into a T, failing
    with "kcpstore: decode object: <err>" when the body does not parse.'
- codeRefs:
  - method:6647c7478f971106b8507d72f97f9d9a
  id: r.get-raw
  level: MUST
  text: 'Resource.GetRaw must GET the named resource in the target namespace and logical
    cluster and return the raw bytes, wrapping a failure as "kcpstore: read <resource>
    <name> in <cluster>: <err>".'
- codeRefs:
  - function:41b4ccba5b43f3162c43c6569fe3adaf
  id: r.host-required
  level: MUST
  text: 'New must reject an Options whose Host is empty with the error "kcpstore:
    Host is required".'
- codeRefs:
  - method:708cca7ea762c4eac4f03c27a2e78121
  id: r.list-cluster
  level: MUST
  text: 'Resource.List must GET the resource across the logical cluster and decode
    the items array, wrapping failure as "kcpstore: list <resource> in <cluster>:
    <err>".'
- codeRefs:
  - method:08d58c09d0e2e942a57a5d0cd38b26bf
  id: r.list-namespace
  level: MUST
  text: 'Resource.ListIn must GET the resource in one namespace of a logical cluster
    and decode the items array, wrapping failure as "kcpstore: list <resource> in
    <cluster>/<namespace>: <err>".'
- codeRefs:
  - file:impl/kcpstore/kcpstore_test.go
  - file:impl/kcpstore/live_test.go
  id: r.live-tests
  level: MAY
  text: The package may be exercised against a real kcp instance through the live
    tests alongside the unit tests.
- codeRefs:
  - method:2da1765147db771d09748666fa261d52
  id: r.mint-token
  level: MUST
  text: Store.MintServiceAccountToken must default an empty namespace to "default",
    POST a TokenRequest with the TTL converted to seconds at the serviceaccounts/<name>/token
    subresource in the core v1 group, and fail when the response carries no token.
- codeRefs:
  - method:c256b1a18aecee599f7d008d7300042e
  id: r.patch-json
  level: MUST
  text: 'Resource.PatchJSON must apply the patch as a JSON patch to the named resource,
    failing with "kcpstore: patch <resource> <name> in <cluster>: <err><detail>".'
- codeRefs:
  - method:060eafce6e59dd5e79f3672d53ae6d3b
  id: r.patch-status-merge
  level: MUST
  text: Resource.PatchStatus must apply the patch at the status subresource as a merge
    patch after statuspatch.WithResourceVersion stamps the target's resource version,
    and must fail with the response detail on error.
- codeRefs:
  - file:impl/kcpstore/paths.go
  - method:67159703f7d37d59b1fec3d3fe2613b0
  id: r.path-cache-definitive-only
  level: MUST
  text: 'PathCache.Lookup must store only what ClusterPath answers definitively: success,
    ErrNoPath, or an IsNotFound error. Every other error must return the empty string
    without writing an entry, so it is asked again.'
- codeRefs:
  - method:2124e46883b2f5ba56788dc76b299726
  - method:f26fdfbd10d0713dc8243504308be1dc
  id: r.path-cache-forget
  level: MUST
  text: PathCache.Forget must drop the entry for a logical cluster so the next Lookup
    resolves it again, and PathCache.Len must report the number of cached entries;
    both must be safe under the cache mutex.
- codeRefs:
  - method:67159703f7d37d59b1fec3d3fe2613b0
  - struct:bd87c7e07b92542841411a9474f66987
  id: r.path-cache-hit-skips-store
  level: MUST
  text: PathCache.Lookup must return a cached path without calling the store when
    the logical cluster is already present, and must tolerate a nil store by caching
    the empty string.
- codeRefs:
  - file:impl/kcpstore/paths.go
  - method:67159703f7d37d59b1fec3d3fe2613b0
  id: r.path-cache-retry-healing-errors
  level: MUST
  text: A refusal, an outage, an unparseable body, and a cancelled call must all be
    treated as non-definitive, because a token mid-refresh and a grant not yet visible
    are refusals that heal; the allowlist form means an unanticipated error is retried
    rather than frozen into a workspace's name.
- codeRefs:
  - method:b56020606bf6fc9651ef87cd37dbe904
  - method:b590040c169d2610823c037f9e6acc73
  id: r.remove-finalizer-absent-ok
  level: MUST
  text: Resource.RemoveFinalizer must read the current finalizers and remove one,
    treating an object that is no longer there as success, and it must delegate the
    patch to RemoveKnownFinalizer.
- codeRefs:
  - method:b56020606bf6fc9651ef87cd37dbe904
  id: r.remove-known-finalizer-noop
  level: MUST
  text: Resource.RemoveKnownFinalizer must build the removal patch from the supplied
    current list and skip the request entirely when statuspatch.FinalizerRemove returns
    a nil patch.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/kcpstore/kcpstore.go` file kcpstore.go (impl/kcpstore/kcpstore.go)
- `file:impl/kcpstore/kcpstore_test.go` file kcpstore_test.go (impl/kcpstore/kcpstore_test.go)
- `file:impl/kcpstore/live_test.go` file live_test.go (impl/kcpstore/live_test.go)
- `file:impl/kcpstore/paths.go` file paths.go (impl/kcpstore/paths.go)
- `function:260af476e14d0dff1cfd24e6619437a0` function NewPathCache (impl/kcpstore/paths.go)
- `function:41b4ccba5b43f3162c43c6569fe3adaf` function New (impl/kcpstore/kcpstore.go)
- `function:96752734b0a49c97d1a14ee1329189da` function Of (impl/kcpstore/kcpstore.go)
- `function:d7d93064d1945db238404c68788f971f` function IsNotFound (impl/kcpstore/kcpstore.go)
- `method:060eafce6e59dd5e79f3672d53ae6d3b` method Resource.PatchStatus (impl/kcpstore/kcpstore.go)
- `method:08d58c09d0e2e942a57a5d0cd38b26bf` method Resource.ListIn (impl/kcpstore/kcpstore.go)
- `method:0adeb1384568f6dec86b61baf285da43` method Resource.Delete (impl/kcpstore/kcpstore.go)
- `method:2124e46883b2f5ba56788dc76b299726` method PathCache.Len (impl/kcpstore/paths.go)
- `method:22f545280ed249562e9ca742d5a77c44` method Store.ClusterPath (impl/kcpstore/kcpstore.go)
- `method:28c264fe91c941c628570295cc91bcea` method Resource.Get (impl/kcpstore/kcpstore.go)
- `method:2da1765147db771d09748666fa261d52` method Store.MintServiceAccountToken (impl/kcpstore/kcpstore.go)
- `method:5763969969c0f4e887f1c499a7fccf73` method Resource.AddFinalizer (impl/kcpstore/kcpstore.go)
- `method:6647c7478f971106b8507d72f97f9d9a` method Resource.GetRaw (impl/kcpstore/kcpstore.go)
- `method:67159703f7d37d59b1fec3d3fe2613b0` method PathCache.Lookup (impl/kcpstore/paths.go)
- `method:708cca7ea762c4eac4f03c27a2e78121` method Resource.List (impl/kcpstore/kcpstore.go)
- `method:b56020606bf6fc9651ef87cd37dbe904` method Resource.RemoveKnownFinalizer (impl/kcpstore/kcpstore.go)
- `method:b590040c169d2610823c037f9e6acc73` method Resource.RemoveFinalizer (impl/kcpstore/kcpstore.go)
- `method:c00fe3e383047902bea23ff36069e91f` method Store.Config (impl/kcpstore/kcpstore.go)
- `method:c256b1a18aecee599f7d008d7300042e` method Resource.PatchJSON (impl/kcpstore/kcpstore.go)
- `method:c311b4200adb69d996ac127afb62ed56` method Store.ClientFor (impl/kcpstore/kcpstore.go)
- `method:cfd282c2aae4ed3dd3a523783c18bc20` method Resource.Finalizers (impl/kcpstore/kcpstore.go)
- `method:e54d2aabb8486cab3ffb3fccd9e528b4` method Resource.Create (impl/kcpstore/kcpstore.go)
- `method:f26fdfbd10d0713dc8243504308be1dc` method PathCache.Forget (impl/kcpstore/paths.go)
- `struct:432c046ab7a1ce292d7fa0372790eed7` struct Store (impl/kcpstore/kcpstore.go)
- `struct:bd87c7e07b92542841411a9474f66987` struct PathCache (impl/kcpstore/paths.go)
- `struct:be2a24917d7a285dc53d560ba36c744a` struct Options (impl/kcpstore/kcpstore.go)
- `struct:dafdd052ffecc9d0a327766098dc06f3` struct Resource (impl/kcpstore/kcpstore.go)
<!-- SPECD_MANAGED_END -->
