# Context: impl-openbaoclient

Repository: `kcp-libs`

This context exists so that the rest of the library can treat OpenBao as one interchangeable PKI backend behind a narrow Go surface. It separates OpenBao's REST and HTTP error vocabulary from the domain types (pki.Health, pki.RootCA, pki.IntermediateCSR, pki.Role, pki.Cert, pki.CertRequest), so consumers never import the OpenBao SDK or inspect status codes, and so the same PKI workflows can be exercised against an in-process test server.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: impl/openbaoclient/openbaoclient.go
  kind: struct
  name: Client
  signature: type Client struct
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.CAChain
  signature: func (c *Client) CAChain(ctx context.Context, namespace, mount string)
    (string, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.CASerial
  signature: func (c *Client) CASerial(ctx context.Context, namespace, mount string)
    (string, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.DeleteNamespace
  signature: func (c *Client) DeleteNamespace(ctx context.Context, path string) error
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.EnsureMount
  signature: func (c *Client) EnsureMount(ctx context.Context, namespace, path, kind
    string) error
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.EnsureNamespace
  signature: func (c *Client) EnsureNamespace(ctx context.Context, path string) error
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.GenerateIntermediate
  signature: func (c *Client) GenerateIntermediate(ctx context.Context, namespace,
    mount, commonName string) (pki.IntermediateCSR, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.GenerateRoot
  signature: func (c *Client) GenerateRoot(ctx context.Context, namespace, mount,
    commonName, ttl string) (pki.RootCA, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.Health
  signature: func (c *Client) Health(ctx context.Context) (pki.Health, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.Issue
  signature: func (c *Client) Issue(ctx context.Context, namespace, mount, role string,
    req pki.CertRequest) (pki.Cert, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.NamespaceExists
  signature: func (c *Client) NamespaceExists(ctx context.Context, path string) (bool,
    error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.SetSignedIntermediate
  signature: func (c *Client) SetSignedIntermediate(ctx context.Context, namespace,
    mount, chain string) error
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.SignIntermediate
  signature: func (c *Client) SignIntermediate(ctx context.Context, rootNamespace,
    mount, csr, commonName, ttl string) (string, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: Client.WriteRole
  signature: func (c *Client) WriteRole(ctx context.Context, namespace, mount, name
    string, role pki.Role) error
- file: impl/openbaoclient/openbaoclient.go
  kind: function
  name: New
  signature: func New(opts Options) (*Client, error)
- file: impl/openbaoclient/openbaoclient.go
  kind: struct
  name: Options
  signature: type Options struct
- file: impl/openbaoclient/openbaoclient.go
  kind: struct
  name: ResponseError
  signature: type ResponseError struct { Method string; Path string; Status int; Body
    string }
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: ResponseError.Error
  signature: func (e *ResponseError) Error() string
- file: impl/openbaoclient/openbaoclient.go
  kind: method
  name: ResponseError.Is
  signature: func (e *ResponseError) Is(target error) bool
requirements:
- codeRefs:
  - function:97f09c010ce8629e3cb888dd53f5fe44
  - struct:644a4d9f9b0e4ee3d1be7476cdc52999
  - struct:d32e9faf15c28f010e9517f571192b92
  id: r.client-construction
  level: MUST
  text: New builds a *Client from Options and returns an error when the options are
    not usable.
- codeRefs:
  - method:61f0984d8204ee9a9d2c470dc79017b6
  id: r.health-mapping
  level: MUST
  text: Health queries sys/health and returns a pki.Health carrying initialized, sealed,
    standby and version; a transport failure is translated into a ResponseError naming
    GET sys/health.
- codeRefs:
  - method:6ff60253ccaf2e7c0441bd69c2ba6e7f
  - method:890d7a7eace7151f12a90cd55188fe44
  - method:9694619ccbdd31bd588b00696f35e46c
  id: r.intermediate-authority
  level: SHOULD
  text: GenerateIntermediate produces a pki.IntermediateCSR for a common name, SignIntermediate
    signs that CSR on a root namespace and mount with a TTL and returns the chain,
    and SetSignedIntermediate installs the chain into the intermediate mount.
- codeRefs:
  - method:9040cc8dcc5c1972e362711cde609457
  id: r.mount-provisioning
  level: MUST
  text: EnsureMount enables a PKI engine at the given namespace and path with the
    given kind, and reports an error when an existing mount has a different type.
- codeRefs:
  - method:039af99f5f709e24f2739149f0f619f7
  - method:279ae14fb9a84351e08e13706a74777a
  - method:308cced5879283d753eb43408c1645ba
  id: r.namespace-lifecycle
  level: MUST
  text: The client can create a namespace, test whether a namespace exists, and delete
    a namespace, each taking a namespace path and reporting a translated error on
    failure.
- codeRefs:
  - method:9355743912f96aae23c89c23f11437c6
  - method:a078027b3f0dfe8153b040d123df1507
  - method:a100c179495b135b3fe2b971348dfedb
  id: r.no-default-issuer-is-no-authority
  level: MUST
  text: An HTTP 400 whose body reports that no default issuer is currently configured
    means the pki mount holds no authority yet, exactly like an HTTP 404, so errors.Is(err,
    pki.ErrNoAuthority) is true for CASerial and for CAChain when the mount exists
    and has no issuer. Every other HTTP 400 stays an error that is neither pki.ErrNoAuthority
    nor ErrNotFound, and its ResponseError still records the method, the path, the
    status and the body.
- codeRefs:
  - file:impl/openbaoclient/openbaoclient_test.go
  id: r.no-default-issuer-tests
  level: MUST
  text: Tests against the in-process OpenBao-shaped server cover both halves. A 400
    whose body says no default issuer is currently configured makes CASerial report
    pki.ErrNoAuthority, and an unrelated 400 does not report pki.ErrNoAuthority and
    does not report ErrNotFound.
- codeRefs:
  - method:bc10c8978afe004cb92eee07347f2fbf
  - struct:e00b5c976e1ed1448875b77d0cec020b
  id: r.response-error-shape
  level: MUST
  text: 'A failed OpenBao request is reported as a *ResponseError that records the
    HTTP method, the request path, the status code and the response body, and Error
    formats them as "openbao: %s /v1/%s -> HTTP %d: %s".'
- codeRefs:
  - method:d06e05f8cc830da5173bb7f340c18d37
  - method:fdf9868ddff367c3a03b7e89abf414b8
  id: r.role-and-issuance
  level: MUST
  text: WriteRole stores a pki.Role under a named role on a mount, and Issue requests
    a certificate for that role from a pki.CertRequest, returning a pki.Cert.
- codeRefs:
  - method:671d53de5144fe4d487f4388a0695574
  - method:9355743912f96aae23c89c23f11437c6
  - method:a078027b3f0dfe8153b040d123df1507
  id: r.root-authority
  level: MUST
  text: GenerateRoot creates a root CA in the given namespace and mount from a common
    name and TTL and returns a pki.RootCA; CASerial and CAChain read back the authority's
    serial number and PEM chain.
- codeRefs:
  - method:a100c179495b135b3fe2b971348dfedb
  - struct:e00b5c976e1ed1448875b77d0cec020b
  id: r.sentinel-error-matching
  level: MUST
  text: ResponseError.Is maps HTTP 404 to ErrNotFound, HTTP 403 to ErrForbidden and
    an HTTP 400 that reports no default issuer as pki.ErrNoAuthority, and returns
    false for every other status, so callers can use errors.Is against the sentinels.
- codeRefs:
  - file:impl/openbaoclient/openbaoclient_test.go
  - file:impl/openbaoclient/testcert_test.go
  id: r.test-coverage
  level: SHOULD
  text: Behavior is exercised by tests against an in-process OpenBao-shaped server,
    including a mount type mismatch reported through ResponseError, with certificate
    fixtures supplied by the test certificate helper.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/openbaoclient/openbaoclient.go` file openbaoclient.go (impl/openbaoclient/openbaoclient.go)
- `file:impl/openbaoclient/openbaoclient_test.go` file openbaoclient_test.go (impl/openbaoclient/openbaoclient_test.go)
- `file:impl/openbaoclient/testcert_test.go` file testcert_test.go (impl/openbaoclient/testcert_test.go)
- `function:50cc797fdaa9ea30e1c2268227d49e6f` function New (impl/openbaoclient/openbaoclient.go)
- `method:028ba10088ea7e1c9b088f1371ff6f1c` method Client.Health (impl/openbaoclient/openbaoclient.go)
- `method:06d62fc43abe23d63deb2db4632e1e1e` method Client.SignIntermediate (impl/openbaoclient/openbaoclient.go)
- `method:1aea693d3bb71792a60e5be463529d7d` method Client.CASerial (impl/openbaoclient/openbaoclient.go)
- `method:2780ee841e94c2842351b25f86a9aba6` method Client.GenerateIntermediate (impl/openbaoclient/openbaoclient.go)
- `method:28c4bfafbaa07d6c82e2084b324aa194` method Client.EnsureNamespace (impl/openbaoclient/openbaoclient.go)
- `method:36ab1558443e4dc608d55d9b54f9a89c` method Client.NamespaceExists (impl/openbaoclient/openbaoclient.go)
- `method:62c6db661c632979b8c3f7e19b7798c1` method Client.SetSignedIntermediate (impl/openbaoclient/openbaoclient.go)
- `method:936472024a18a6ae02d83697145811e6` method Client.Issue (impl/openbaoclient/openbaoclient.go)
- `method:a100c179495b135b3fe2b971348dfedb` method ResponseError.Is (impl/openbaoclient/openbaoclient.go)
- `method:bc10c8978afe004cb92eee07347f2fbf` method ResponseError.Error (impl/openbaoclient/openbaoclient.go)
- `method:d434c58e3c847579feb579420d450ba3` method Client.GenerateRoot (impl/openbaoclient/openbaoclient.go)
- `method:daa5ec84247fd7f6efc20c9ed1ab66a0` method Client.EnsureMount (impl/openbaoclient/openbaoclient.go)
- `method:e68f0e43ec497ceea63eadb4056e8953` method Client.CAChain (impl/openbaoclient/openbaoclient.go)
- `method:e958540130931c8d8324f14876bb1e0b` method Client.WriteRole (impl/openbaoclient/openbaoclient.go)
- `method:edc1b59fa40dbebe402cfff04b5aa69c` method Client.DeleteNamespace (impl/openbaoclient/openbaoclient.go)
- `struct:312293f0f3d4b547b434febccf5f9ab3` struct Client (impl/openbaoclient/openbaoclient.go)
- `struct:e00b5c976e1ed1448875b77d0cec020b` struct ResponseError (impl/openbaoclient/openbaoclient.go)
- `struct:e3b96d9757623d14e22d2a1091966b68` struct Options (impl/openbaoclient/openbaoclient.go)
<!-- SPECD_MANAGED_END -->
