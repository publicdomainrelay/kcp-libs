# Context: impl-pkiprovisioner

Repository: `kcp-libs`

This context exists so that callers can ask for a root CA, a per-namespace intermediate authority, and leaf certificates without knowing the underlying PKI store's call sequence. It owns the ordering rules — mount the root, generate or read back the root, create the namespace, mount PKI there, sign and install an intermediate, write the role — and it owns the caching that keeps those steps from running on every request. It is the implementation behind the pki.Provisioner interface declared in abc/pki.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/pkiprovisioner/pkiprovisioner.go` file pkiprovisioner.go (impl/pkiprovisioner/pkiprovisioner.go)
- `file:impl/pkiprovisioner/pkiprovisioner_test.go` file pkiprovisioner_test.go (impl/pkiprovisioner/pkiprovisioner_test.go)
- `function:30364e141ea2faa4b4234973c4a19b7f` function New (impl/pkiprovisioner/pkiprovisioner.go)
- `function:d3c3c16dc7874b597b0ada7624dae197` function LeafChain (impl/pkiprovisioner/pkiprovisioner.go)
- `method:4195ab23df62ee1db3840769986974c8` method Provisioner.EnsureRoot (impl/pkiprovisioner/pkiprovisioner.go)
- `method:4e00b7939435016777ed505a3c634185` method Provisioner.Issue (impl/pkiprovisioner/pkiprovisioner.go)
- `method:7aaa1b393f6662ecc89ab29299cbea68` method Provisioner.CachedRootPEM (impl/pkiprovisioner/pkiprovisioner.go)
- `method:d5967140fe7eb80411ee9ed7cc877c96` method Provisioner.Delete (impl/pkiprovisioner/pkiprovisioner.go)
- `method:dbce55e518a5f94002a8102e3348772c` method Provisioner.EnsureAuthority (impl/pkiprovisioner/pkiprovisioner.go)
- `struct:85d52f39514274817830fa863a880ebb` struct Provisioner (impl/pkiprovisioner/pkiprovisioner.go)
- `struct:d27d806e5c22258e2ce7cad7037a9dda` struct Options (impl/pkiprovisioner/pkiprovisioner.go)
<!-- SPECD_MANAGED_END -->
