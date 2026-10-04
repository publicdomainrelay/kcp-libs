# open-architecture/kcp-libs

The architecture and specs of `kcp-libs`, as kcp holds them.

Written by specd, one commit per change in kcp. An orphan branch: it shares no
history with the code and is never checked out next to it, so an agent working
on the code never mistakes the spec for the code.

| path | holds |
| --- | --- |
| `arch.yaml` | the whole repository as generated system contexts (`kind: GeneratedArchitecture`) |
| `repository.yaml` | the Repository manifest, its populate state, and the commits every context shares |
| `specs/<context>.yaml` | each context's declared spec; edit here to change kcp |
| `status/<context>.yaml` | observed code facts and conditions |
| `context/<context>.md` | the context's prose and resolved code references; the spec lives in `specs/` |
| `changes/<name>.yaml` | each SpecChange: direction, delta, progress, outcome |
| `CHANGES.md` | on a feature branch: the requirement delta against the default branch |
| `graph/*.jsonl` | the context graph, one vertex or edge per line |
