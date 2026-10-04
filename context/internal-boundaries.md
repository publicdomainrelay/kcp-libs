# Context: internal-boundaries

Repository: `kcp-libs`

This context exists to keep the kcp-libs module's ABC-style layering honest by test rather than by convention. The tests read the real import graph from the Go toolchain, so any new import that crosses a layer boundary, points at internal test support from production code, or pulls a module-local package into the common leaf layer fails the build. The spec records the invariants the test file enforces and the vocabulary (layers and their ranks) it relies on, so the boundaries package can be understood without re-deriving the rules from the Go source.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
