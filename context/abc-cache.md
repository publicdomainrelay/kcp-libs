# Context: abc-cache

Repository: `kcp-libs`

This context exists so a reader can see what the cache promises without reading the implementation: which index contract an object store must satisfy, how a Set fans a lookup out over several indexers of the same kind, and which field extractors the library offers for objects it did not define. It also fixes the indexer set that IndexersFor derives from the configured labels, which is the seam the surrounding reconcile and informer layers plug into.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/cache/cache.go` file cache.go (abc/cache/cache.go)
- `file:abc/cache/cache_test.go` file cache_test.go (abc/cache/cache_test.go)
- `function:323ee6d71bcc903db5a96ba67a3e1292` function RefOf (abc/cache/cache.go)
- `function:368cfc3b7333cde3fc16c28b5f4a251e` function NewSet (abc/cache/cache.go)
- `function:3b4476e3e6d3931e3e5f66d88938a0d2` function PhaseOf (abc/cache/cache.go)
- `function:53a9fed74460c17bc06521cc3e273e8c` function NameOf (abc/cache/cache.go)
- `function:5ce1c083464f112da22046a1de3a5518` function IndexersFor (abc/cache/cache.go)
- `function:7b1ca2f92d1acfeb4fc9be87c387a001` function ClusterOf (abc/cache/cache.go)
- `function:ac84053b66aaf34b0c83c98bd076de38` function Decode (abc/cache/cache.go)
- `function:cf99478a03753478fb8bd3cc621df402` function NestedString (abc/cache/cache.go)
- `interface:83657a0a0b8c546e00b0f6ddd68c8e0a` interface Indexer (abc/cache/cache.go)
- `method:23f07591814f8acee78325cb538a547d` method Set.Names (abc/cache/cache.go)
- `method:2f035c41d933e105b559d662a9a88767` method Indexer.ByIndex (abc/cache/cache.go)
- `method:3cf5981edb352b4ae0f722e3ed6bb96c` method Indexer.GetByKey (abc/cache/cache.go)
- `method:436e0dc34ebdec584138d188650d491e` method Set.ByIndex (abc/cache/cache.go)
- `method:79db186239dd9ebf94db73d3c1834051` method Set.List (abc/cache/cache.go)
- `method:7dedfffcc434baa8546c274ac5fa4131` method Set.Get (abc/cache/cache.go)
- `method:995ab025f452ec7f8d6ac24f85f6bbba` method Set.Add (abc/cache/cache.go)
- `method:e8d2e2d111f92694f5202ff04b380d0c` method Indexer.List (abc/cache/cache.go)
- `struct:b3ef85326db0e635df938cf47a312813` struct Set (abc/cache/cache.go)
- `type_alias:a2a398908517249bfded29f79dbe609e` type_alias Indexers (abc/cache/cache.go)
- `type_alias:c324eb8aebee5659b8c3144e047c30a2` type_alias IndexFunc (abc/cache/cache.go)
<!-- SPECD_MANAGED_END -->
