# Context: impl-assets

Repository: `kcp-libs`

This context exists so callers can hand a declarative set of asset files to one object and get back a stable name-to-path map, without caring when or how often the bytes hit disk. It separates one-shot materialisation, which is idempotent and retried after failure, from explicit rewriting, which always writes, and it gives the DNS workload a ready-made set for its shim and probe scripts.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/assets/assets.go` file assets.go (impl/assets/assets.go)
- `file:impl/assets/assets_test.go` file assets_test.go (impl/assets/assets_test.go)
- `file:impl/assets/dns.go` file dns.go (impl/assets/dns.go)
- `file:impl/assets/dns_test.go` file dns_test.go (impl/assets/dns_test.go)
- `file:impl/assets/dnsprobe.ts` file dnsprobe.ts (impl/assets/dnsprobe.ts)
- `file:impl/assets/dnsshim.ts` file dnsshim.ts (impl/assets/dnsshim.ts)
- `function:d3ebda2faf6210986c5ca5687c578339` function DNSSet (impl/assets/dns.go)
- `method:459575a1ced2aec2dc09df6c846496d1` method Set.Materialise (impl/assets/assets.go)
- `method:621b8207809a46a163d2af9949761ea0` method Set.Rewrite (impl/assets/assets.go)
- `method:dfdb88e0a88513d22977521f62fadc3c` method Set.Path (impl/assets/assets.go)
- `struct:6a6e6a1cca0d3f0928da5e266ee9e2da` struct Set (impl/assets/assets.go)
<!-- SPECD_MANAGED_END -->
