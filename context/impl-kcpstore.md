# Context: impl-kcpstore

Repository: `kcp-libs`

This context exists so that consumers (controllers, admission, DNS, servicenames) have one place that knows how to reach kcp: URL shape, per-cluster client reuse, rate limiting, JSON decoding, status and finalizer patching, token minting, and workspace path resolution. The path cache is the newest part and it is deliberately conservative — it exists because a workspace path that failed to resolve must not be frozen into a wrong name, so only an answer that cannot change by asking again is stored.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:impl/kcpstore/kcpstore.go` file kcpstore.go (impl/kcpstore/kcpstore.go)
- `file:impl/kcpstore/kcpstore_test.go` file kcpstore_test.go (impl/kcpstore/kcpstore_test.go)
- `file:impl/kcpstore/live_test.go` file live_test.go (impl/kcpstore/live_test.go)
- `file:impl/kcpstore/paths.go` file paths.go (impl/kcpstore/paths.go)
- `function:260af476e14d0dff1cfd24e6619437a0` function NewPathCache (impl/kcpstore/paths.go)
- `function:41b4ccba5b43f3162c43c6569fe3adaf` function New (impl/kcpstore/kcpstore.go)
- `function:96752734b0a49c97d1a14ee1329189da` function Of (impl/kcpstore/kcpstore.go)
- `function:d7d93064d1945db238404c68788f971f` function IsNotFound (impl/kcpstore/kcpstore.go)
- `method:060eafce6e59dd5e79f3672d53ae6d3b` method Resource.PatchStatus (impl/kcpstore/kcpstore.go)
- `method:08d58c09d0e2e942a57a5d0cd38b26bf` method Resource.ListIn (impl/kcpstore/kcpstore.go)
- `method:0adeb1384568f6dec86b61baf285da43` method Resource.Delete (impl/kcpstore/kcpstore.go)
- `method:2124e46883b2f5ba56788dc76b299726` method PathCache.Len (impl/kcpstore/paths.go)
- `method:22f545280ed249562e9ca742d5a77c44` method Store.ClusterPath (impl/kcpstore/kcpstore.go)
- `method:28c264fe91c941c628570295cc91bcea` method Resource.Get (impl/kcpstore/kcpstore.go)
- `method:2da1765147db771d09748666fa261d52` method Store.MintServiceAccountToken (impl/kcpstore/kcpstore.go)
- `method:5763969969c0f4e887f1c499a7fccf73` method Resource.AddFinalizer (impl/kcpstore/kcpstore.go)
- `method:6647c7478f971106b8507d72f97f9d9a` method Resource.GetRaw (impl/kcpstore/kcpstore.go)
- `method:67159703f7d37d59b1fec3d3fe2613b0` method PathCache.Lookup (impl/kcpstore/paths.go)
- `method:708cca7ea762c4eac4f03c27a2e78121` method Resource.List (impl/kcpstore/kcpstore.go)
- `method:b56020606bf6fc9651ef87cd37dbe904` method Resource.RemoveKnownFinalizer (impl/kcpstore/kcpstore.go)
- `method:b590040c169d2610823c037f9e6acc73` method Resource.RemoveFinalizer (impl/kcpstore/kcpstore.go)
- `method:c00fe3e383047902bea23ff36069e91f` method Store.Config (impl/kcpstore/kcpstore.go)
- `method:c256b1a18aecee599f7d008d7300042e` method Resource.PatchJSON (impl/kcpstore/kcpstore.go)
- `method:c311b4200adb69d996ac127afb62ed56` method Store.ClientFor (impl/kcpstore/kcpstore.go)
- `method:cfd282c2aae4ed3dd3a523783c18bc20` method Resource.Finalizers (impl/kcpstore/kcpstore.go)
- `method:e54d2aabb8486cab3ffb3fccd9e528b4` method Resource.Create (impl/kcpstore/kcpstore.go)
- `method:f26fdfbd10d0713dc8243504308be1dc` method PathCache.Forget (impl/kcpstore/paths.go)
- `struct:432c046ab7a1ce292d7fa0372790eed7` struct Store (impl/kcpstore/kcpstore.go)
- `struct:bd87c7e07b92542841411a9474f66987` struct PathCache (impl/kcpstore/paths.go)
- `struct:be2a24917d7a285dc53d560ba36c744a` struct Options (impl/kcpstore/kcpstore.go)
- `struct:dafdd052ffecc9d0a327766098dc06f3` struct Resource (impl/kcpstore/kcpstore.go)
<!-- SPECD_MANAGED_END -->
