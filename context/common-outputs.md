# Context: common-outputs

Repository: `kcp-libs`

This context exists to fix the contract for converting arbitrary JSON-ish values into the string map that kcp surfaces as outputs. The two functions are the single place where that conversion rule lives, so the exec runner and the policy client cannot drift apart on how a number, boolean, or list becomes a string. The spec records the observable contract: nil in gives nil out, strings pass through unchanged, non-strings become their JSON encoding, and a value JSON cannot encode still yields a usable string rather than an error.

_Write the prose above and the fields in the spec block. `codeRefs` and the resolved references below are maintained by the tool; an edit there is lost._

## spec

```yaml spec
interfaces:
- file: common/outputs/outputs.go
  kind: function
  name: Stringify
  signature: func Stringify(in map[string]any) map[string]string
- file: common/outputs/outputs.go
  kind: function
  name: StringifyValue
  signature: func StringifyValue(value any) string
requirements:
- codeRefs:
  - file:common/outputs/outputs_test.go
  - function:d16f38628fc00d680c6f54309865e821
  id: r.behavior-is-tested
  level: SHOULD
  text: TestStringify covers nil input, a plain string, a number, a boolean, and a
    list, pinning the documented conversions.
- codeRefs:
  - file:common/outputs/outputs.go
  - function:d16f38628fc00d680c6f54309865e821
  id: r.empty-input-gives-nil
  level: MUST
  text: Stringify returns nil, not an empty map, when the input map is nil or has
    length zero.
- codeRefs:
  - function:1f7915d8cd9aa1f3759d833f5af90128
  - function:d16f38628fc00d680c6f54309865e821
  id: r.every-entry-stringified
  level: MUST
  text: Stringify returns a map holding the same keys as the input, where each value
    is the result of StringifyValue applied to the input value.
- codeRefs:
  - function:1f7915d8cd9aa1f3759d833f5af90128
  id: r.marshal-failure-falls-back
  level: MUST
  text: StringifyValue returns fmt.Sprint of the value when json.Marshal fails, so
    the function never returns an error and never panics on an unencodable value.
- codeRefs:
  - file:common/outputs/outputs.go
  - function:1f7915d8cd9aa1f3759d833f5af90128
  id: r.non-strings-json-encoded
  level: MUST
  text: StringifyValue returns the JSON encoding of a value that is not a string,
    so numbers become bare numerals, booleans become true or false, and lists become
    compact JSON arrays.
- codeRefs:
  - file:common/outputs/outputs.go
  - function:1f7915d8cd9aa1f3759d833f5af90128
  id: r.strings-pass-through
  level: MUST
  text: StringifyValue returns a value whose dynamic type is string unchanged, without
    quoting or escaping it.
upstream: self
```

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

- `file:common/outputs/outputs.go` file outputs.go (common/outputs/outputs.go)
- `file:common/outputs/outputs_test.go` file outputs_test.go (common/outputs/outputs_test.go)
- `function:1f7915d8cd9aa1f3759d833f5af90128` function StringifyValue (common/outputs/outputs.go)
- `function:d16f38628fc00d680c6f54309865e821` function Stringify (common/outputs/outputs.go)
<!-- SPECD_MANAGED_END -->
