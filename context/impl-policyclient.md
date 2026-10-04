# Context: impl-policyclient

Repository: `kcp-libs`

This context exists to give the rest of the repository one place that speaks the policy engine's HTTP protocol, so callers depend on the abc/policy.Client interface and never build requests themselves. It exists because policy runs are submitted as JSON workflows to an engine endpoint and their verdicts come back as task status, and both halves need the same validation, error wrapping and output flattening.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/policyclient/policyclient.go` file policyclient.go (impl/policyclient/policyclient.go)
- `file:impl/policyclient/policyclient_test.go` file policyclient_test.go (impl/policyclient/policyclient_test.go)
- `function:54f3e1e2bd8e260021c98693f9a2c90a` function Outputs (impl/policyclient/policyclient.go)
- `function:b98ae160f1fe55260dca95e09986a5bb` function PolicyName (impl/policyclient/policyclient.go)
- `function:ceea74205a5754163b309ccbb838c89a` function New (impl/policyclient/policyclient.go)
- `method:19f667c165b3c1a5fd0e639d57f02e89` method Client.Submit (impl/policyclient/policyclient.go)
- `method:4e934ce27e912f8b62f60030591215cc` method Client.Status (impl/policyclient/policyclient.go)
- `struct:cadc1b79f4d841a6641b1b7ed9b4d3ec` struct Client (impl/policyclient/policyclient.go)
<!-- SPECD_MANAGED_END -->
