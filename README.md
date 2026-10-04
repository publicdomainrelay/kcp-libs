# open-architecture/kcp-libs

The architecture and specs of `kcp-libs`, as kcp holds them.

Written by specd, one commit per change in kcp. An orphan branch: it shares no
history with the code and is never checked out next to it, so an agent working
on the code never mistakes the spec for the code.

| path | holds |
| --- | --- |
| `arch.yaml` | the whole repository as open architecture system contexts |
| `repository.yaml` | the Repository manifest and its populate state |
| `specs/<context>.yaml` | each context's declared spec; edit here to change kcp |
| `status/<context>.yaml` | observed code facts and conditions |
| `context/<context>.md` | the CLM context document a model reads and edits |
| `changes/<name>.yaml` | each SpecChange: direction, delta, progress, outcome |
| `graph/*.jsonl` | the context graph, one vertex or edge per line |
