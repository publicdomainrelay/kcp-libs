# Context: examples-pki

Repository: `kcp-libs`

This context exists to document the examples/pki example: an executable end-to-end demonstration that provisions a PKI setup through KCP and issues certificates, backed by an in-process fake Vault. The fake Vault exists so the example and its test can run without a real Vault server, while still recording every HTTP call for assertions.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
