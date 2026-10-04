# Context: abc-store

Repository: `kcp-libs`

This context exists to fix the storage contract that the rest of kcp-libs codes against. Controllers, caches and reconcilers need to read and write cluster objects and mint service account tokens, but they must not depend on a concrete Kubernetes client or on a specific workload runtime. abc-store supplies that seam as small generic interfaces plus one pure comparison helper. Unchanged exists because status patches should only be sent when the object actually changed on the wire, and that comparison must follow JSON encoding rules rather than Go struct identity.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: abc/store/store.go
  kind: interface
  name: Reader
  signature: type Reader[T any] interface { Get(ctx context.Context, r ref.Ref) (*T,
    error); List(ctx context.Context, logicalCluster string) ([]T, error) }
- file: abc/store/store.go
  kind: method
  name: Reader.Get
  signature: Get(ctx context.Context, r ref.Ref) (*T, error)
- file: abc/store/store.go
  kind: method
  name: Reader.List
  signature: List(ctx context.Context, logicalCluster string) ([]T, error)
- file: abc/store/store.go
  kind: interface
  name: Resource
  signature: type Resource interface{}
- file: abc/store/store.go
  kind: interface
  name: TokenMinter
  signature: type TokenMinter interface { MintServiceAccountToken(ctx context.Context,
    logicalCluster, namespace, name string, ttl time.Duration) (string, error) }
- file: abc/store/store.go
  kind: method
  name: TokenMinter.MintServiceAccountToken
  signature: MintServiceAccountToken(ctx context.Context, logicalCluster, namespace,
    name string, ttl time.Duration) (string, error)
- file: abc/store/store.go
  kind: function
  name: Unchanged
  signature: func Unchanged(a, b any) bool
- file: abc/store/store.go
  kind: interface
  name: Writer
  signature: type Writer[T any] interface { Create(ctx context.Context, logicalCluster
    string, obj *T) error; Delete(ctx context.Context, r ref.Ref) error; PatchStatus(ctx
    context.Context, r ref.Ref, patch []byte) error; RemoveFinalizer(ctx context.Context,
    r ref.Ref, finalizer string) error }
- file: abc/store/store.go
  kind: method
  name: Writer.Create
  signature: Create(ctx context.Context, logicalCluster string, obj *T) error
- file: abc/store/store.go
  kind: method
  name: Writer.Delete
  signature: Delete(ctx context.Context, r ref.Ref) error
- file: abc/store/store.go
  kind: method
  name: Writer.PatchStatus
  signature: PatchStatus(ctx context.Context, r ref.Ref, patch []byte) error
- file: abc/store/store.go
  kind: method
  name: Writer.RemoveFinalizer
  signature: RemoveFinalizer(ctx context.Context, r ref.Ref, finalizer string) error
requirements:
- codeRefs:
  - file:abc/store/store.go
  - interface:24a75b1e85b321842f9208ef68c555dd
  - interface:35f2d3c152cb806dafbc20869a2b1645
  id: r.generic-reader-writer-pair
  level: SHOULD
  text: Reader and Writer should stay generic over T so a single pair of interfaces
    serves every stored object type, and should name no concrete API server or workload
    runtime.
- codeRefs:
  - interface:35f2d3c152cb806dafbc20869a2b1645
  - method:dd7423535cde997b6ddc93c5ec334c4b
  - method:e52cadf103fd9081db57993476b63580
  id: r.reader-get-and-list
  level: MUST
  text: Reader must expose Get, which takes a context and a ref.Ref and returns a
    pointer to T, and List, which takes a context and a logicalCluster string and
    returns a slice of T.
- codeRefs:
  - file:abc/store/store.go
  - interface:9d5d8b965e6b6d107bc98f738bd212ac
  id: r.resource-marker
  level: MUST
  text: Resource must be a generic interface that marks a type as storable by the
    Reader and Writer implementations.
- codeRefs:
  - interface:60623c60df0bde883df69e0a6fd8778d
  - method:1d478157f0c21fec12ff68cca5afa4d8
  id: r.token-minter
  level: MUST
  text: TokenMinter must expose MintServiceAccountToken, which takes a context, logicalCluster,
    namespace, name and ttl, and returns the minted token string or an error.
- codeRefs:
  - file:abc/store/store.go
  - file:abc/store/store_test.go
  - function:989c182ec2114a55ce5c64d494875860
  id: r.unchanged-compares-wire-shape
  level: MUST
  text: Unchanged must marshal both arguments with encoding/json and report equality
    by comparing the encoded bytes, so that two values with equal wire shape are equal
    regardless of Go struct identity or map key order.
- codeRefs:
  - file:abc/store/store_test.go
  - function:989c182ec2114a55ce5c64d494875860
  id: r.unchanged-fails-closed-on-encode-error
  level: MUST
  text: Unchanged must return false when either argument fails to marshal to JSON,
    so an unencodable value is never reported as unchanged.
- codeRefs:
  - interface:24a75b1e85b321842f9208ef68c555dd
  - method:0c58430aa89b1920a3744d460c85b616
  - method:b66a329085f8ae0870f5487662772dd4
  - method:cf43c7d659eb7321032b9dd58fe618c7
  - method:fbb58fe6aca56610baea0a33ce6e3d5f
  id: r.writer-mutations
  level: MUST
  text: Writer must expose Create taking a context, logicalCluster and object pointer;
    Delete taking a context and ref.Ref; PatchStatus taking a context, ref.Ref and
    a raw patch byte slice; and RemoveFinalizer taking a context, ref.Ref and the
    finalizer name. Each returns an error.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/store/store.go` file store.go (abc/store/store.go)
- `file:abc/store/store_test.go` file store_test.go (abc/store/store_test.go)
- `function:989c182ec2114a55ce5c64d494875860` function Unchanged (abc/store/store.go)
- `interface:24a75b1e85b321842f9208ef68c555dd` interface Writer (abc/store/store.go)
- `interface:35f2d3c152cb806dafbc20869a2b1645` interface Reader (abc/store/store.go)
- `interface:60623c60df0bde883df69e0a6fd8778d` interface TokenMinter (abc/store/store.go)
- `interface:9d5d8b965e6b6d107bc98f738bd212ac` interface Resource (abc/store/store.go)
- `method:0c58430aa89b1920a3744d460c85b616` method Writer.Create (abc/store/store.go)
- `method:1d478157f0c21fec12ff68cca5afa4d8` method TokenMinter.MintServiceAccountToken (abc/store/store.go)
- `method:b66a329085f8ae0870f5487662772dd4` method Writer.Delete (abc/store/store.go)
- `method:cf43c7d659eb7321032b9dd58fe618c7` method Writer.RemoveFinalizer (abc/store/store.go)
- `method:dd7423535cde997b6ddc93c5ec334c4b` method Reader.List (abc/store/store.go)
- `method:e52cadf103fd9081db57993476b63580` method Reader.Get (abc/store/store.go)
- `method:fbb58fe6aca56610baea0a33ce6e3d5f` method Writer.PatchStatus (abc/store/store.go)
<!-- SPECD_MANAGED_END -->
