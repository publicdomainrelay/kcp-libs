# Context: impl-pkiprovisioner

Repository: `kcp-libs`

This context exists so that callers can ask for a root CA, a per-namespace intermediate authority, and leaf certificates without knowing the underlying PKI store's call sequence. It owns the ordering rules — mount the root, generate or read back the root, create the namespace, mount PKI there, sign and install an intermediate, write the role — and it owns the caching that keeps those steps from running on every request. It is the implementation behind the pki.Provisioner interface declared in abc/pki.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
