# Context: abc-runref

Repository: `kcp-libs`

This context exists so a reconciler can tell a fresh start from a duplicate of a run that is already in flight. Records live only for the configured TTL, so the index is a cache of recent starts rather than durable state, and callers are expected to refresh a record on each pass (Keep) and drop it when the run reaches a terminal phase. AlreadyStarted encodes the refusal rule the admission path needs, including the cases that must NOT be refused: no record, the same runID, a run that already started, a retry, or a recreated object that carries a different UID behind the same name.

_The resolved code references are regenerated on every run. Cite the ids above rather than writing them here._

<!-- SPECD_MANAGED_BEGIN -->
## Resolved code references

_None yet._
<!-- SPECD_MANAGED_END -->
