# Context: examples-pki

Repository: `kcp-libs`

This context exists to document the examples/pki example: an executable end-to-end demonstration that provisions a PKI setup through KCP and issues certificates, backed by an in-process fake Vault. The fake Vault exists so the example and its test can run without a real Vault server, while still recording every HTTP call for assertions.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: examples/pki/main.go
  kind: function
  name: Run
  signature: (ctx context.Context, out io.Writer) error
- file: examples/pki/vault.go
  kind: method
  name: vault.CallCount
  signature: () int
- file: examples/pki/vault.go
  kind: method
  name: vault.Close
  signature: () error
- file: examples/pki/vault.go
  kind: method
  name: vault.ServeHTTP
  signature: (w http.ResponseWriter, r *http.Request)
- file: examples/pki/vault.go
  kind: method
  name: vault.URL
  signature: () string
requirements:
- codeRefs:
  - file:examples/pki/certs.go
  id: r.cert-helpers
  level: SHOULD
  text: Certificate handling for the example SHOULD live in certs.go and stay separate
    from the entry point and the fake vault.
- codeRefs:
  - file:examples/pki/main_test.go
  id: r.pki-tested
  level: SHOULD
  text: The example SHOULD stay covered by main_test.go, which runs it against the
    fake vault.
- codeRefs:
  - file:examples/pki/main.go
  - function:ee6adf675850825fe2a3b3696b991879
  id: r.run-entry-point
  level: MUST
  text: The example MUST expose Run, which takes a context and an output writer and
    returns an error, as its executable entry point.
- codeRefs:
  - file:examples/pki/vault.go
  - method:864fe0fd004ec6e68a16c123dd516ef4
  id: r.vault-call-count
  level: MUST
  text: The fake vault MUST expose CallCount, returning the number of calls it has
    recorded, so a caller can assert on observed traffic.
- codeRefs:
  - file:examples/pki/vault.go
  - method:aae965e31b6d34cc2cc7c5d18d44f176
  id: r.vault-close
  level: MUST
  text: The fake vault MUST expose Close, returning an error, so its HTTP server is
    shut down after use.
- codeRefs:
  - file:examples/pki/vault.go
  - method:147b7070b45be7c13a88225e5f5c7e37
  id: r.vault-serves-vault-api
  level: MUST
  text: The fake vault MUST implement ServeHTTP over http.ResponseWriter and *http.Request
    so it can act as an HTTP handler for the Vault API paths the example calls.
- codeRefs:
  - file:examples/pki/vault.go
  - method:1a848c7f40482ad066fc74dd3ef4c787
  id: r.vault-url
  level: MUST
  text: The fake vault MUST expose URL, returning the base URL the example uses to
    reach it.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:examples/pki/certs.go` file certs.go (examples/pki/certs.go)
- `file:examples/pki/main.go` file main.go (examples/pki/main.go)
- `file:examples/pki/main_test.go` file main_test.go (examples/pki/main_test.go)
- `file:examples/pki/vault.go` file vault.go (examples/pki/vault.go)
- `function:ee6adf675850825fe2a3b3696b991879` function Run (examples/pki/main.go)
- `method:147b7070b45be7c13a88225e5f5c7e37` method vault.ServeHTTP (examples/pki/vault.go)
- `method:1a848c7f40482ad066fc74dd3ef4c787` method vault.URL (examples/pki/vault.go)
- `method:864fe0fd004ec6e68a16c123dd516ef4` method vault.CallCount (examples/pki/vault.go)
- `method:aae965e31b6d34cc2cc7c5d18d44f176` method vault.Close (examples/pki/vault.go)
<!-- SPECD_MANAGED_END -->
