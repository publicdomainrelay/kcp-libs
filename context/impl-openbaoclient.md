# Context: impl-openbaoclient

Repository: `kcp-libs`

This context exists so that the rest of the library can treat OpenBao as one interchangeable PKI backend behind a narrow Go surface. It separates OpenBao's REST and HTTP error vocabulary from the domain types (pki.Health, pki.RootCA, pki.IntermediateCSR, pki.Role, pki.Cert, pki.CertRequest), so consumers never import the OpenBao SDK or inspect status codes, and so the same PKI workflows can be exercised against an in-process test server.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
