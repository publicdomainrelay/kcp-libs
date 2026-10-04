# Context: impl-openbaoclient

Repository: `kcp-libs`

This context exists so that the rest of the library can treat OpenBao as one interchangeable PKI backend behind a narrow Go surface. It separates OpenBao's REST and HTTP error vocabulary from the domain types (pki.Health, pki.RootCA, pki.IntermediateCSR, pki.Role, pki.Cert, pki.CertRequest), so consumers never import the OpenBao SDK or inspect status codes, and so the same PKI workflows can be exercised against an in-process test server.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

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
