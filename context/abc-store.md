# Context: abc-store

Repository: `kcp-libs`

This context exists to fix the storage contract that the rest of kcp-libs codes against. Controllers, caches and reconcilers need to read and write cluster objects and mint service account tokens, but they must not depend on a concrete Kubernetes client or on a specific workload runtime. abc-store supplies that seam as small generic interfaces plus one pure comparison helper. Unchanged exists because status patches should only be sent when the object actually changed on the wire, and that comparison must follow JSON encoding rules rather than Go struct identity.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:abc/store/store.go` file store.go (abc/store/store.go)
- `file:abc/store/store_test.go` file store_test.go (abc/store/store_test.go)
- `function:989c182ec2114a55ce5c64d494875860` function Unchanged (abc/store/store.go)
- `interface:24a75b1e85b321842f9208ef68c555dd` interface Writer (abc/store/store.go)
- `interface:35f2d3c152cb806dafbc20869a2b1645` interface Reader (abc/store/store.go)
- `interface:60623c60df0bde883df69e0a6fd8778d` interface TokenMinter (abc/store/store.go)
- `interface:9d5d8b965e6b6d107bc98f738bd212ac` interface Resource (abc/store/store.go)
- `method:0c58430aa89b1920a3744d460c85b616` method Writer.Create (abc/store/store.go)
- `method:1d478157f0c21fec12ff68cca5afa4d8` method TokenMinter.MintServiceAccountToken (abc/store/store.go)
- `method:b66a329085f8ae0870f5487662772dd4` method Writer.Delete (abc/store/store.go)
- `method:cf43c7d659eb7321032b9dd58fe618c7` method Writer.RemoveFinalizer (abc/store/store.go)
- `method:dd7423535cde997b6ddc93c5ec334c4b` method Reader.List (abc/store/store.go)
- `method:e52cadf103fd9081db57993476b63580` method Reader.Get (abc/store/store.go)
- `method:fbb58fe6aca56610baea0a33ce6e3d5f` method Writer.PatchStatus (abc/store/store.go)
<!-- SPECD_MANAGED_END -->
