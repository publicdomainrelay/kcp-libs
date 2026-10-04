# Changes on `open-architecture/kcp-libs--fix-openbao-no-default-issuer`

The requirement-level delta against `open-architecture/kcp-libs`, and what this branch realized.

## Requirements

### abc-cache

- intent: "" -> "This context exists so a reader can see what the cache promises without reading the implementation: which index contract an object store must satisfy, how a Set fans a lookup out over several indexers of the same kind, and which field extractors the library offers for objects it did not define. It also fixes the indexer set that IndexersFor derives from the configured labels, which is the seam the surrounding reconcile and informer layers plug into."
- added `r.byindex-skips-errors` (MUST): "Set.ByIndex concatenates ByIndex results across the indexers of a kind and skips any indexer that returns an error, so one failing store cannot fail the whole lookup."
- added `r.decode-to-type` (SHOULD): "Decode turns an unstructured object into a typed value, returning an error when the conversion fails."
- added `r.get-first-hit` (MUST): "Set.Get asks the indexers of a kind in registered order and returns the first object whose GetByKey reports exists, with false when no indexer has it."
- added `r.indexer-contract` (MUST): "An Indexer offers exactly three reads: GetByKey for one key with an exists flag and an error, ByIndex for the objects carrying a value under a named index, and List for all objects."
- added `r.list-concatenates` (MUST): "Set.List returns the concatenation of List from every indexer registered under a kind."
- added `r.multi-indexer-registry` (MUST): "A Set holds a list of indexers per object kind. Set.Add appends an Indexer under a kind, so several stores can back the same kind."
- added `r.names-from-index` (MUST): "Set.Names returns the names of the objects found at an index value through Set.ByIndex, using NameOf on each object."
- added `r.nested-field-walk` (MUST): "NestedString walks a list of nested field names on an object and returns the string found there."
- added `r.object-field-extractors` (MUST): "The package reads common fields off unstructured objects: ClusterOf returns the cluster, RefOf returns a ref.Ref with a found flag, NameOf returns the name and PhaseOf returns the phase."
- added `r.standard-indexers` (MUST): "IndexersFor builds an Indexers map from a parent label, a job label and a trigger pod field, so callers get the standard index set from configuration."
- added `r.tested-against-fake-indexer` (SHOULD): "The Set behaviour is exercised in abc/cache/cache_test.go against a fake indexer whose ByIndex uses IndexersFor with the policy workflow pod label, the job run label and the policyWorkflowPod field."

### abc-joballoc

- intent: "" -> "This context exists so the abc layer has one place that answers which names a job already holds and which of those a controller has not yet seen, without each consumer keeping its own bookkeeping. It is deliberately runtime-agnostic: it names no workload runtime and depends only on common/ref and common/expiringmap, so any reconciler in the repo can use it. Expiry comes from the shared expiringmap rather than from a timer here, which keeps the allocator free of goroutines and makes every read testable against an explicit now."
- added `r.allocate-appends-and-skips-empty` (MUST): "Allocator.Allocate returns immediately when names is empty, and otherwise reads the existing allocations for the ref and appends one Allocation{Name, At: now} per name before setting the result back."
- added `r.forget-deletes-the-job` (MUST): "Allocator.Forget deletes the ref's entry from the underlying map."
- added `r.len-reports-held-jobs` (MUST): "Allocator.Len returns the number of entries the underlying map currently holds."
- added `r.merge-names-order-preserving-dedupe` (MUST): "MergeNames concatenates the given groups in order, dropping empty names and names already emitted, so the result keeps first-seen order with no duplicates."
- added `r.names-nil-when-absent-or-empty` (MUST): "Allocator.Names returns nil when the ref has no live entry or the entry holds no allocations, and otherwise returns the allocation names in stored order."
- added `r.new-wires-expiring-map-with-ttl` (MUST): "New returns an Allocator whose byJob map is an expiringmap.Map keyed by ref.Ref holding []Allocation, created with the given ttl."
- added `r.pending-excludes-observed-and-dedupes` (MUST): "Allocator.Pending returns nil when the ref has no live entry, and otherwise returns the allocated names not present in observed, skipping any name already emitted so the result carries no duplicates."
- added `r.reads-honour-explicit-now` (MUST): "Every read and write takes an explicit now time.Time and passes it to the expiringmap, so expiry is decided by the caller's clock and not by the allocator."

### abc-pki

- intent: "" -> "This context exists to fix the boundary between the reconciler/provisioner code and the certificate authority backend. It names the operations a PKI backend must offer in terms of namespaces, mounts, roles and certificate issuance, so that an OpenBao-backed implementation (impl/openbaoclient) and a caching, idempotent wrapper (impl/pkiprovisioner) can be swapped without the consumers changing. The data carriers mirror the JSON field names the backend returns, which lets the implementations decode responses directly."
- added `r.generate-root` (MUST): "The Client interface MUST let a caller generate a root CA from namespace, mount, common name and TTL, returning a RootCA carrying certificate and serial."
- added `r.health-shape` (SHOULD): "Health SHOULD describe a backend as Initialized, Sealed, Standby and Version, with the first three as bool and Version as string."
- added `r.intermediate-pair` (MUST): "The Client interface MUST support the intermediate flow as a pair: GenerateIntermediate returns an IntermediateCSR with CSR and private key, and SignIntermediate takes the root namespace, mount, CSR, common name and TTL and returns the signed chain."
- added `r.issue-certificate` (MUST): "The Client interface MUST issue a certificate from namespace, mount, role and a CertRequest (CommonName, AltNames, IPSANs, TTL), returning a Cert with Certificate, PrivateKey, IssuingCA, CAChain and Serial."
- added `r.json-tags` (SHOULD): "Cert, RootCA, IntermediateCSR and Health SHOULD keep their JSON tags (certificate, private_key, issuing_ca, ca_chain, serial_number, csr, initialized, sealed, standby, version) so responses decode without a mapping layer."
- added `r.mount-kind` (MUST): "The Client interface MUST let a caller ensure a mount inside a namespace, taking the namespace, the path and the mount kind."
- added `r.namespace-lifecycle` (MUST): "The Client interface MUST let a caller create a namespace (EnsureNamespace) and remove a namespace (DeleteNamespace), each addressed by path and returning error."
- added `r.package-defaults` (MUST): "The package MUST define DefaultMount "pki", DefaultRole "denopod", DefaultRootCommonName "kcp-mesh-root", DefaultRootTTL "87600h", DefaultIntermediateTTL "43800h" and DefaultLeafTTL "720h" as constants."
- added `r.provisioner-cached-root-pem` (MUST): "Provisioner.CachedRootPEM MUST return the root certificate as bytes with no error and no context, so callers can read it without an authority round trip."
- added `r.provisioner-delete` (MUST): "Provisioner.Delete MUST take a path and return error, so a reconciled path can be torn down."
- added `r.provisioner-ensure-authority` (MUST): "Provisioner.EnsureAuthority MUST take a path and return an Authority carrying Namespace, CommonName, Serial and Chain."
- added `r.provisioner-ensure-root` (MUST): "Provisioner.EnsureRoot MUST return the RootCA for the provisioner's own root without the caller naming namespace or mount."
- added `r.provisioner-issue` (MUST): "Provisioner.Issue MUST take a path, common name, alternative names and IPs and return a Cert, hiding mount, role and TTL from the caller."
- added `r.provisioner-surface` (MUST): "The Provisioner interface MUST offer EnsureRoot, EnsureAuthority, Issue, Delete and CachedRootPEM as its whole method set."
- added `r.read-ca-serial-and-chain` (MUST): "The Client interface MUST expose the CA serial and the CA chain of a mount as separate string reads (CASerial, CAChain)."
- added `r.sentinel-no-authority` (MUST): "The package MUST export ErrNoAuthority, an error value meaning the namespace holds no authority yet."
- added `r.set-signed-intermediate` (MUST): "The Client interface MUST let a caller install a signed intermediate chain into namespace and mount (SetSignedIntermediate)."
- added `r.write-role` (MUST): "The Client interface MUST let a caller write a named Role into a namespace and mount; Role carries AllowedDomains, AllowSubdomains, AllowBareDomains, EnforceHostnames, KeyType, KeyBits and MaxTTL."

### abc-policy

- intent: "" -> "This context exists so that consumers of a policy engine depend on a stable, transport-free contract instead of a concrete HTTP client. It defines the vocabulary and signatures that an implementation such as impl/policyclient must satisfy, and the shape of the task status that callers poll after submission. Keeping the contract in abc/policy lets the engine endpoint stay a per-call argument, so one Client talks to many policy engines."
- added `r.client-contract` (MUST): "The package MUST export a Client interface whose Submit call returns a task identifier string and an error, and whose Status call returns a Task and an error."
- added `r.state-vocabulary` (MUST): "The package MUST define exactly three run states as string constants: StateRunning with value "running", StateSucceeded with value "succeeded", and StateFailed with value "failed"."
- added `r.status-lookup` (MUST): "Status MUST identify a run by the engine endpoint plus the task identifier returned from Submit, and MUST accept a context for cancellation."
- added `r.submit-inputs` (MUST): "Submit MUST take the engine endpoint as a string, the workflow as raw JSON bytes, and the run inputs as a map[string]string, and MUST accept a context for cancellation."
- added `r.task-shape` (MUST): "Task MUST carry State, ExitStatus, Outputs as a map[string]string, and Message, all as plain string or map fields."

### abc-probe

- intent: "" -> "This context exists to specify the consecutive-failure probe abstraction: a shared, concurrency-safe place where repeated observations about a named probe key accumulate into a trip/no-trip decision. It is deliberately narrow, holding only counting and threshold normalization, so that callers decide what a key means and what tripping it should do. The behavior encoded here, that a success resets the count, that a new run ID restarts the count, and that keys are independent, is what the tests pin down."
- added `r.counter-records-run-and-failures` (MUST): "A Counter must carry the RunID of the workload instance it describes and the number of consecutive Failures observed for it."
- added `r.effective-threshold-defaults-to-one` (MUST): "EffectiveThreshold must return the supplied threshold when it is positive and fall back to a default of one when it is zero or negative."
- added `r.forget-drops-a-key` (MUST): "Tracker.Forget must remove the counter held for the given key, whether or not that key is currently present."
- added `r.keys-are-independent` (MUST): "Each probe key must keep its own counter, so failures recorded under one key never affect the count or the trip decision of another key."
- added `r.len-reports-key-count` (MUST): "Tracker.Len must report the number of keys currently held by the tracker."
- added `r.new-run-restarts-count` (MUST): "Tracker.Record must detect a run ID that differs from the stored one and restart that key's count from zero, so a restarted workload is not judged by its predecessor's failures."
- added `r.newtracker-returns-empty-tracker` (MUST): "NewTracker must return a Tracker with an initialized, empty counter map, ready for use without further setup."
- added `r.record-trips-at-threshold` (MUST): "Tracker.Record must increment the failure count for the key on a failed observation and return true only when the count reaches the threshold, so failures below the threshold return false."
- added `r.success-resets-count` (MUST): "Tracker.Record must reset the consecutive failure count to zero on a passed observation and return false, so the next failure starts counting from one again."
- added `r.tracker-guards-its-map` (MUST): "Tracker must hold its counters in a map keyed by the probe key string and guard every access with a mutex, so Record, Forget and Len are safe to call concurrently."

### abc-queue

- intent: "" -> "This context exists to specify the admission/lease contract of the abc queue: what a Run must look like to be planned, how Policy defaults resolve, how a Capacity, Blocker and reserved count combine into a per-run Admission, and how held leases are scoped to a parent, counted against observed lifecycle phase, expired by TTL and released. It is the durable description of the queue's pure decision functions and its lease bookkeeping, so consumers such as the deno process runner and the admission factory can rely on the same semantics without re-deriving them from the source."
- added `r.admission-waiting` (MUST): "Admission must report through Waiting whether the run it describes is still waiting for capacity rather than admitted."
- added `r.decision-verdict` (MUST): "Decision must take a policy, an optional maxConcurrent override and the active and ahead counts, and return whether the run is admitted together with the reason strings that explain the verdict."
- added `r.leases-grant-forget-len` (MUST): "Leases must expose Grant to record a run under a parent at a given time, Forget to drop the entry for a run, and Len to report the number of live entries."
- added `r.leases-parent-scoped-count` (MUST): "Leases.Count must count only the leases granted against the given parent ref, so a lease held for one parent is never counted for another."
- added `r.leases-release-on-observed` (MUST): "A lease must be released when the run it covers is present in the observed map and its phase, as read through the Lifecycle, is running or terminal, so that Count returns zero and Len drops the entry once both runs are observed."
- added `r.leases-ttl-expiry` (MUST): "Leases must be created through NewLeases with a TTL, and every entry must expire once that TTL has elapsed so that an expired lease no longer contributes to the count."
- added `r.limit-resolution` (MUST): "Limit must resolve a policy plus an optional maxConcurrent override into an effective concurrency limit and a flag saying whether that limit is unlimited."
- added `r.plan-admissions` (MUST): "Plan must take the ordered runs together with capacity, an optional blocker, the reserved count and the lifecycle, and return one Admission per run, and PlanIndex must expose the same result keyed by ref.Ref."
- added `r.policy-defaulting` (MUST): "Policy must be an alias resolved to usable defaults by EffectivePolicy, so a caller can pass a partially set or zero policy and receive a complete one."
- added `r.queue-purity` (SHOULD): "The queue decision functions should stay pure and workload-agnostic: they take runs, capacity, policy and lifecycle as arguments and name no workload runtime."
- added `r.run-ordering` (MUST): "Order must sort a slice of runs in place into the queue's admission order before planning."
- added `r.wake-list` (MUST): "WakeList must return the refs of the runs that should be woken given the runs, the lifecycle, and either a limit or the unlimited flag."

### abc-reconcile

- intent: "" -> "This context exists to separate the decision half of a controller from the transport and runtime half. A Reconciler is pure: given the observed object it returns what it wants to happen, with no client, no workqueue and no dependency on any workload runtime. A Handler wraps the other direction, owning retries and terminality so the reconcile loop can be scheduled; Policy centralises the backoff arithmetic instead of scattering it through call sites, and Bridge is the piece that actually reads, decides and applies against a cluster. Result and Patch exist so the decided outcome can be accumulated, interrogated and emitted as a patch without the reconciler having to know how the write is performed. The package names no workload runtime so the generic layers of the library stay usable without the deno packages."
- added `r.bridge-drives-a-handler` (SHOULD): "Bridge must present the Handler shape so a caller can read, decide and apply against a cluster through one Process call."
- added `r.handler-reports-outcome` (MUST): "Handler.Process must take a context and a Key and report the delay before the next attempt, whether the key is terminal, and any error, so the reconcile loop can schedule the next run."
- added `r.handlerfunc-adapts-a-function` (MUST): "HandlerFunc must adapt a plain function with the Process shape to the Handler interface."
- added `r.key-identifies-an-object` (MUST): "Key must carry a Kind string and a ref.Ref, and Key.String must render as Kind joined by "/" to Ref.Key()."
- added `r.patch-renders-a-result` (SHOULD): "Patch must turn a Result[Status] into a merge patch, so a decided outcome can be emitted without the reconciler knowing how the write happens."
- added `r.policy-owns-backoff` (MUST): "Policy must own the retry arithmetic: Default gives the starting delay, Clamp bounds a delay to the policy range, Next combines the current delay, the key and terminality into the next delay and whether to continue, and ConflictAfter gives the delay used after a conflict."
- added `r.reconciler-decides-from-observed` (MUST): "Reconciler.Reconcile must take a context and the observed object and return a Result[Status] with an error, keeping the decision free of any client or runtime dependency."
- added `r.reconcilerfunc-adapts-a-function` (MUST): "ReconcilerFunc must adapt a plain function with the Reconcile shape to the Reconciler interface."
- added `r.result-accumulates-operations` (MUST): "Result must record the decided operations by Kind: Add records a kind without a target, AddFor records a kind against a ref.Ref target, and Has reports whether a kind was recorded."
- added `r.result-exposes-cleanup-obligations` (MUST): "Result must expose the cleanup side of the decision: Deletes reports whether the object is being deleted, ReleasesFinalizer reports whether the finalizer can be removed, and IsCleared reports whether a named status field was cleared."

### abc-runner

- intent: "" -> "This context exists to fix the boundary between the generic workload lifecycle and its concrete implementations. The package names no runtime beyond the field vocabulary of PodRequest, so execrunner, memoryrunner and the examples can implement or consume the same four verbs without the reconciliation, queue and watch layers depending on any one runner. The spec records that contract: the two interfaces, the status and request shapes they exchange, and the State alias both statuses share."
- added `r.engine-request-status` (MUST): "EngineRequest and EngineStatus must exist as distinct types from the pod pair, so an engine run is described and reported without reusing the pod request and status shapes."
- added `r.engine-runner-lifecycle` (MUST): "EngineRunner must define the four lifecycle verbs Start, Observe, Stop and Probe, in that order, mirroring PodRunner for engine-shaped workloads."
- added `r.engine-runner-observe` (MUST): "EngineRunner.Observe must accept a context and the run identifier and return EngineStatus together with an error."
- added `r.engine-runner-probe` (MUST): "EngineRunner.Probe must accept a context, the run identifier, a command argv slice and a timeout, and return a boolean plus an error, matching the PodRunner probe shape exactly."
- added `r.engine-runner-start` (MUST): "EngineRunner.Start must accept a context and EngineRequest and return the run identifier as a string together with an error."
- added `r.engine-runner-stop` (MUST): "EngineRunner.Stop must accept a context and the run identifier and return only an error."
- added `r.pod-request-fields` (MUST): "PodRequest must carry Name, LogicalCluster, DenoJSON, DenoLock, Script, PermissionArgs, Env, Token, Server and Workspace, so a pod run receives both its Deno module inputs and its identity and placement context."
- added `r.pod-runner-lifecycle` (MUST): "PodRunner must define the four lifecycle verbs Start, Observe, Stop and Probe, in that order, so a pod implementation can be started, inspected, stopped and probed through one interface."
- added `r.pod-runner-observe` (MUST): "PodRunner.Observe must accept a context and the run identifier and return PodStatus together with an error."
- added `r.pod-runner-probe` (MUST): "PodRunner.Probe must accept a context, the run identifier, a command argv slice and a timeout, and return a boolean plus an error."
- added `r.pod-runner-start` (MUST): "PodRunner.Start must accept a context and PodRequest and return the run identifier as a string together with an error."
- added `r.pod-runner-stop` (MUST): "PodRunner.Stop must accept a context and the run identifier and return only an error."
- added `r.pod-status-fields` (MUST): "PodStatus must carry State, ExitCode as an int32, Message and an Outputs map of string to string."
- added `r.state-alias` (MUST): "State must be a string alias, so that status types can report a lifecycle phase as a plain string value."

### abc-runref

- intent: "" -> "This context exists so a reconciler can tell a fresh start from a duplicate of a run that is already in flight. Records live only for the configured TTL, so the index is a cache of recent starts rather than durable state, and callers are expected to refresh a record on each pass (Keep) and drop it when the run reaches a terminal phase. AlreadyStarted encodes the refusal rule the admission path needs, including the cases that must NOT be refused: no record, the same runID, a run that already started, a retry, or a recreated object that carries a different UID behind the same name."
- added `r.already-started-no-record-no-refusal` (MUST): "AlreadyStarted returns false when found is false, when known.RunID is empty, or when known.RunID equals current.RunID."
- added `r.already-started-refuses-stale-pending-copy` (MUST): "AlreadyStarted returns true only for a found record with a different non-empty runID against a current run that has not started, has no retries, and has no conflicting UID; this is the refusal of a stale Pending copy."
- added `r.already-started-retry-not-duplicate` (MUST): "AlreadyStarted returns false when current.Started is true or current.Retries > 0, so a run that has started before is retrying, not duplicating."
- added `r.already-started-uid-mismatch-not-duplicate` (MUST): "AlreadyStarted returns false when both known.UID and current.UID are non-empty and differ, so a recreated object behind the same name is not refused."
- added `r.forget-deletes-by-normalized-ref` (MUST): "Index.Forget deletes every entry whose Ref equals ref.New(r.LogicalCluster, r.Namespace, r.Name), regardless of runID."
- added `r.keep-terminal-or-empty-runid-forgets` (MUST): "Index.Keep calls Forget when terminal is true or runID is empty, and Record otherwise, so terminal runs leave no trace in the index."
- added `r.len-reports-entry-count` (MUST): "Index.Len returns the number of live entries held by the underlying map."
- added `r.lookup-returns-first-matching-ref` (MUST): "Index.Lookup returns the first entry whose Ref equals the normalized key and true, or the zero Record and false when no entry matches."
- added `r.new-builds-index-with-ttl` (MUST): "New(ttl) returns an Index backed by an expiring map with that TTL, so records age out without an explicit delete."
- added `r.record-ignores-empty-runid` (MUST): "Index.Record with an empty runID is a no-op; nothing is stored."
- added `r.record-normalizes-ref-key` (MUST): "Index.Record stores the entry under the runID, with Ref rebuilt from ref.New(r.LogicalCluster, r.Namespace, r.Name), so the map key and the stored Ref are the normalized triple."
- added `r.record-prunes-other-run-for-same-ref` (MUST): "Index.Record prunes any stored entry whose id differs from the new runID and whose Ref equals the new normalized key, so one ref keeps only the newest run."

### abc-store

- intent: "" -> "This context exists to fix the storage contract that the rest of kcp-libs codes against. Controllers, caches and reconcilers need to read and write cluster objects and mint service account tokens, but they must not depend on a concrete Kubernetes client or on a specific workload runtime. abc-store supplies that seam as small generic interfaces plus one pure comparison helper. Unchanged exists because status patches should only be sent when the object actually changed on the wire, and that comparison must follow JSON encoding rules rather than Go struct identity."
- added `r.generic-reader-writer-pair` (SHOULD): "Reader and Writer should stay generic over T so a single pair of interfaces serves every stored object type, and should name no concrete API server or workload runtime."
- added `r.reader-get-and-list` (MUST): "Reader must expose Get, which takes a context and a ref.Ref and returns a pointer to T, and List, which takes a context and a logicalCluster string and returns a slice of T."
- added `r.resource-marker` (MUST): "Resource must be a generic interface that marks a type as storable by the Reader and Writer implementations."
- added `r.token-minter` (MUST): "TokenMinter must expose MintServiceAccountToken, which takes a context, logicalCluster, namespace, name and ttl, and returns the minted token string or an error."
- added `r.unchanged-compares-wire-shape` (MUST): "Unchanged must marshal both arguments with encoding/json and report equality by comparing the encoded bytes, so that two values with equal wire shape are equal regardless of Go struct identity or map key order."
- added `r.unchanged-fails-closed-on-encode-error` (MUST): "Unchanged must return false when either argument fails to marshal to JSON, so an unencodable value is never reported as unchanged."
- added `r.writer-mutations` (MUST): "Writer must expose Create taking a context, logicalCluster and object pointer; Delete taking a context and ref.Ref; PatchStatus taking a context, ref.Ref and a raw patch byte slice; and RemoveFinalizer taking a context, ref.Ref and the finalizer name. Each returns an error."

### common-clientlimit

- intent: "" -> "This context exists because client-go defaults every client to 5 qps with a burst of 10, which starves a controller that watches and reconciles many objects. The package centralises the fix in one function so each controller construction path (exportwatch, informerwatch, kcpstore) applies the same rate-limit policy without duplicating the defaulting logic, and keeps caller intent: an operator who explicitly sets QPS or Burst on a config is not overridden. It sits in the generic common layer, names no workload runtime, and is usable by any component that builds a rest.Config."
- added `r.caller-value-wins` (MUST): "Apply must set out.QPS only when the copied config's QPS is <= 0 and out.Burst only when its Burst is <= 0, so a rate limit the caller already set on the config is kept."
- added `r.default-above-client-go-default` (MUST): "DefaultQPS must stay greater than 5, because client-go defaults every client to 5 qps with a burst of 10, which starves a controller."
- added `r.explicit-option-overrides-empty-config` (MUST): "When the config carries no rate limit, a positive qps and burst argument must be used as the resulting QPS and Burst."
- added `r.generic-layer-no-runtime` (SHOULD): "The package must stay generic: it names no workload runtime and works from a plain *rest.Config, so any controller construction path can use it."
- added `r.non-positive-option-falls-back-to-default` (MUST): "A qps less than or equal to zero must fall back to DefaultQPS, and a burst less than or equal to zero must fall back to DefaultBurst."
- added `r.returns-copy-not-mutation` (MUST): "Apply must copy the given config with rest.CopyConfig and return the copy; the config passed by the caller must not be mutated."

### common-condition

- intent: "" -> "This context exists to specify the shared condition helpers that kcp-libs controllers use when reporting status. Rather than each controller reimplementing condition bookkeeping, common/condition centralises it: a single Set entry point that records type, status, reason, message and observed generation through meta.SetStatusCondition, thin SetTrue/SetFalse wrappers over that entry point, and the small reader/eraser functions Of, Is, Remove and Copy that callers need around it. It exists so condition handling is uniform, so the observed generation is never forgotten, and so callers can hand out copies of a condition list without aliasing the live one."
- added `r.copy-independent-clone` (MUST): "Copy must return an independent clone of the input condition slice, so that mutating the returned slice does not change the input."
- added `r.is-status-test` (MUST): "Is must report whether the condition of the given kind exists in the list and carries the given status."
- added `r.mutations-through-pointer` (SHOULD): "The mutating helpers Set, SetTrue, SetFalse and Remove should take a pointer to the condition slice and update the caller's list in place rather than returning a new list."
- added `r.observed-generation-recorded` (MUST): "Set must record the supplied generation as the condition's ObservedGeneration, so a condition written at generation 3 reports ObservedGeneration 3."
- added `r.of-lookup-or-nil` (MUST): "Of must return a pointer to the condition of the given kind in the list, and nil when no condition of that kind is present."
- added `r.remove-by-kind` (MUST): "Remove must delete the condition of the given kind from the list, after which Of for that kind returns nil."
- added `r.set-delegates-to-metav1` (MUST): "Set must write a metav1.Condition carrying the given type, status, reason, message and observed generation, by delegating to meta.SetStatusCondition on the caller's list pointer."
- added `r.set-false-shorthand` (MUST): "SetFalse must set the condition with status metav1.ConditionFalse, forwarding the generation, kind, reason and message to Set."
- added `r.set-replaces-same-type` (MUST): "A set must replace the previous condition of the same type, so the list holds at most one condition per type and its length does not grow on repeated sets."
- added `r.set-true-shorthand` (MUST): "SetTrue must set the condition with status metav1.ConditionTrue, forwarding the generation, kind, reason and message to Set."
- added `r.unchanged-condition-keeps-transition-time` (MUST): "Setting a condition whose type, status, reason and message are unchanged must preserve the existing transition time rather than resetting it."

### common-denocomputer

- intent: "" -> "This context exists to keep one copy of the deno.computer names and phases that both this generic library and the deno-kcp consumer agree on. Rather than letting the library invent its own group, finalizers, labels, conditions or phase strings, the constants live here and a test reads the consumer's own source as the source of truth and fails on any drift. The phase predicates sit beside the Phase type so callers in the reconcile paths can ask whether a run, pod, job, policy workflow or policy engine has finished without re-encoding which phases count as terminal in every call site."
- added `r.consumer-checkout-optional` (SHOULD): "The consumer-agreement tests should skip when the deno-kcp checkout is not next to this module, and should fail instead of skipping when KCP_LIBS_REQUIRE_CONSUMER is 1."
- added `r.finalizers-match-consumer` (MUST): "Every finalizer string this module declares (FinalizerDenoRun, FinalizerDenoPod, FinalizerPolicyEngine, FinalizerPolicyWorkflowRun, FinalizerOpenBao) must be one the consumer declares; none may be invented."
- added `r.group-and-version-match-consumer` (MUST): "Group, Version and the derived APIVersion must equal the GroupName and Version the deno-kcp consumer declares in its api/v1alpha1 package."
- added `r.labels-match-consumer` (MUST): "PolicyWorkflowPodLabel, JobRunLabel and TriggerLabel must match the consumer's declared values; TriggerLabel may instead be accepted when the consumer names that exact literal anywhere in its Go source."
- added `r.phase-is-string-alias` (MUST): "Phase is a string type, so phase constants and stored phase strings interchange through a plain string conversion."
- added `r.phases-conditions-reasons-match-consumer` (MUST): "The phase, condition and reason strings this module declares must be the ones the consumer stores: DenoRunPending/Running/Succeeded/Failed, PolicyWorkflowCancelled, RunTriggerTriggered, the five Condition names, the two OpenBao conditions, and the three reason strings ReasonPolicyWorkflowPodMissing, ReasonEngineNotReady and ReasonQueued."
- added `r.running-phase` (MUST): "RunningPhase(phase) must hold exactly when the phase is Running, and must not hold for a finished phase such as Succeeded."
- added `r.terminal-deno-job` (MUST): "TerminalDenoJob(phase) must hold for Succeeded and Failed, and must not hold for a run in flight such as Running or Pending."
- added `r.terminal-deno-pod` (MUST): "TerminalDenoPod(phase) must hold for Succeeded and Failed, and must not hold for a run in flight such as Running or Pending."
- added `r.terminal-deno-run` (MUST): "TerminalDenoRun(phase) must hold for Succeeded and Failed, and must not hold for Running or Pending."
- added `r.terminal-policy-engine` (MUST): "TerminalPolicyEngine(phase) must hold only for Failed; it must not hold for Running or for any other phase."
- added `r.terminal-policy-workflow` (MUST): "TerminalPolicyWorkflow(phase) must hold for Succeeded, Failed and Cancelled, so a cancelled policy run counts as terminal, unlike a cancelled plain run or pod."

### common-denospec

- intent: "" -> "This context exists so that the deno workload vocabulary lives in one Go package that both the KCP libraries and the Deno-side consumer agree on. The structs are the wire contract; the contract test fails if the Go shapes drift from what the consumer declares. Validate and Args exist so that a permission set declared as data becomes a checked, runnable Deno command line rather than an ad-hoc string built at each call site, and EffectiveRestartPolicy and ProbeCommand pin down the remaining runtime decisions around restarts and probes."
- added `r.args-render-permission-flags` (MUST): "Permissions.Args must render the permission set into a Deno argument list covering every flag form, returning an error rather than a partial list when the permissions do not validate."
- added `r.effective-restart-policy-resolves-default` (MUST): "EffectiveRestartPolicy must resolve an unset or unrecognised RestartPolicy to the policy the workload is actually restarted under, so callers need not special-case the empty value."
- added `r.exec-probe-shape-matches-consumer` (MUST): "ExecProbe must carry the exec probe shape the consumer declares, so a probe written on either side decodes on the other."
- added `r.no-workload-runtime-assumption-elsewhere` (SHOULD): "The deno vocabulary should stay in this package, so the generic library layers keep naming no workload runtime and remain usable without deno."
- added `r.permission-allow-and-deny-forms` (MUST): "Permission must express both the bare allow and deny booleans and their AllowList and DenyList member lists, so a permission can be granted or refused wholesale or narrowed to named members."
- added `r.permissions-cover-deno-capabilities` (MUST): "Permissions must cover Deno's capability set: All, NoPrompt, HRTime, and the Read, Write, Net, Env, Run, FFI, Sys, Import and IgnoreEnv permissions, plus AllowScripts for the named scripts that may run."
- added `r.pod-template-describes-workload` (MUST): "PodTemplate must describe the pod the deno workload runs in, including its probe, so a consumer declares the pod in the same shape the Go side decodes."
- added `r.probe-command-is-issued-through-the-shim` (MUST): "ProbeCommand must build the probe invocation from the runtime, shim, probe, name and path, passing only the environment names in allowEnv, so a probe runs under the same runtime and shim as the workload with a bounded environment."
- added `r.service-account-ref-names-identity` (MUST): "ServiceAccountRef must carry the service account Name and an optional Namespace, so a workload can name the identity it runs as without pinning a namespace."
- added `r.shapes-match-consumer-declaration` (MUST): "Every exported shape in this package must carry exactly the field names the TypeScript consumer declares for it, so a value written from one side decodes as the other; the contract test fails when a shape here is invented or drifts."
- added `r.validate-rejects-unrepresentable-permissions` (MUST): "Permissions.Validate must return an error for permission values that cannot be rendered as Deno flags, rejecting commas and empty values, and return nil for a permission set that can be rendered."

### common-expiringmap

- intent: "" -> "This context exists as the shared expiry primitive for the wider kcp-libs project: callers that need a bounded-lifetime cache (for example the job allocator's lookup tables) get one place that owns the TTL comparison, the lazy delete on read, and the locking discipline, rather than each site re-deriving `at.Sub(entry.At) >= ttl` and remembering to hold a lock while doing it. Passing `at` in rather than calling time.Now keeps the expiry behaviour deterministic and testable, which is exactly how the accompanying test file drives it."
- added `r.behaviour-covered-by-tests` (SHOULD): "The package's tests cover set-then-get, the exact-TTL boundary with the expired entry dropped, the default TTL for zero and negative inputs, Delete and DeleteIf counts, Expire sweeping only entries past their TTL, and Range visiting every entry and leaving the map unchanged when the visitor stops."
- added `r.delete-unconditional` (MUST): "Delete removes the key regardless of whether the entry is live or expired, and does nothing when the key is absent."
- added `r.deleteif-predicate-count` (MUST): "DeleteIf drops every entry whose key and value satisfy the caller's predicate and returns the number of entries dropped."
- added `r.entry-stores-value-and-write-time` (MUST): "An Entry holds the stored value together with the time it was written, so expiry is decided against the write time and not against a value the caller supplies later."
- added `r.expire-sweeps-past-ttl` (MUST): "Expire drops every entry that is past its TTL as of the given time and returns the number of entries dropped, leaving entries written more recently than the TTL untouched."
- added `r.expiry-boundary-inclusive` (MUST): "An entry is expired when the elapsed time since its write time is greater than or equal to the TTL, so an entry read at exactly `write + TTL` is already gone."
- added `r.get-drops-expired-and-reports-miss` (MUST): "Get returns the value and true for a live entry, and returns the zero value and false for a missing key or an expired entry; an expired entry found on the way is deleted from the map so the read leaves the map smaller."
- added `r.len-counts-stored-entries` (MUST): "Len returns the number of entries currently stored, so it counts entries that have passed their TTL but have not yet been read or swept."
- added `r.map-type-parameters-and-ttl-field` (MUST): "Map is generic over a comparable key and any value, and carries its TTL as an exported field alongside an unexported mutex and the entry map."
- added `r.mutex-guards-every-operation` (MUST): "Every map operation that touches the entry map holds the mutex, so concurrent Set, Get, Delete, DeleteIf, Expire, Range, SetPruning and Len calls are safe."
- added `r.newmap-defaults-non-positive-ttl` (MUST): "NewMap returns a Map with the given TTL, but any TTL that is zero or negative falls back to defaultTTL; the returned map always has a non-nil entries map."
- added `r.range-snapshot-then-visit` (MUST): "Range copies the live keys and values under the lock, releases the lock, then visits the snapshot; a nil-returning visitor stops the walk early, and stopping or visiting never mutates the map."
- added `r.set-records-write-time` (MUST): "Set stores the value under the key with the caller-supplied time as the entry's write time, overwriting any previous entry for that key."
- added `r.setpruning-sweeps-then-stores` (MUST): "SetPruning drops every entry that is either past its TTL as of the given time or accepted by the non-nil prune predicate, then stores the new key and value under the same lock, so the pruning and the insert are atomic."

### common-kcp

- intent: "" -> "This context exists so the naming rules that map a KCP logical cluster path onto DNS labels and fully qualified service names stay specified in one place. The rest of the repository, the examples and the service-name factory all depend on the exact string shape produced here, so the reversal order, the root-workspace and empty-segment skipping, the default namespace fallback and the omission of an empty label suffix must hold exactly as the tests assert them."
- added `r.fqdn-default-namespace` (MUST): "ServiceFQDN MUST substitute "default" when the namespace argument is empty, and MUST prefix the namespace to the service name as name + "." + namespace."
- added `r.fqdn-empty-labels-omitted` (MUST): "ServiceFQDN MUST append the ServiceLabels result as a dotted suffix only when it is non-empty, so a root-only cluster yields no label segment."
- added `r.fqdn-known-shapes` (MUST): "ServiceFQDN MUST return the asserted shapes: name "pds" with cluster "root:alice" gives "pds.default.alice.svc.kcp.local" under DefaultServiceDomain, and cluster "root" gives "pds.default.svc.kcp.local"."
- added `r.fqdn-svc-domain-suffix` (MUST): "ServiceFQDN MUST terminate the host with the ServiceSegment value "svc" followed by the domain, producing host + ".svc." + domain."
- added `r.labels-root-only-empty` (MUST): "ServiceLabels MUST return the empty string for a logical cluster that contains only the root workspace, for example "root"."
- added `r.labels-skip-empty-and-root` (MUST): "ServiceLabels MUST skip parts that are empty and parts equal to RootWorkspace, and MUST join the remaining parts with "."."
- added `r.labels-split-on-colon` (MUST): "ServiceLabels MUST split the logical cluster on ":" and walk the resulting parts from the last to the first, so the deepest workspace segment appears first in the result."
- added `r.tests-cover-helpers` (SHOULD): "Table-driven tests SHOULD keep covering ServiceLabels and ServiceFQDN, including the root-only and empty-namespace cases."

### common-logging

- intent: "" -> "This context exists so that structured logging has one definition instead of every package building its own slog handler. It fixes the decisions that would otherwise drift: output format is JSON, the zero Level means Info rather than Debug, and the service name is a top-level attribute rather than embedded in messages. Discard exists so tests and quiet paths can be handed a real *slog.Logger with no output at all, which keeps signatures uniform and avoids nil-logger branches in callers."
- added `r.discard-writes-nothing` (MUST): "Discard must return a *slog.Logger whose handler writes to io.Discard at slog.LevelError, so callers get a usable logger that emits no output."
- added `r.new-defaults-level-to-info` (MUST): "New must replace a zero opts.Level with slog.LevelInfo, so an unset level means Info and not the zero value of slog.Level."
- added `r.new-tags-records-with-the-service` (MUST): "New must attach a "service" attribute to the logger only when opts.Service is not empty; an empty Service must leave the logger untagged."
- added `r.new-writes-json-to-the-writer` (MUST): "New must build the logger on a slog.NewJSONHandler over opts.Writer with the resolved level, so records are JSON lines on the writer the caller supplied."
- added `r.options-is-the-only-wiring` (SHOULD): "Loggers should be produced only through Options via New or Discard, so format, default level, and the service attribute stay in one place."

### common-outputs

- intent: "" -> "This context exists to fix the contract for converting arbitrary JSON-ish values into the string map that kcp surfaces as outputs. The two functions are the single place where that conversion rule lives, so the exec runner and the policy client cannot drift apart on how a number, boolean, or list becomes a string. The spec records the observable contract: nil in gives nil out, strings pass through unchanged, non-strings become their JSON encoding, and a value JSON cannot encode still yields a usable string rather than an error."
- added `r.behavior-is-tested` (SHOULD): "TestStringify covers nil input, a plain string, a number, a boolean, and a list, pinning the documented conversions."
- added `r.empty-input-gives-nil` (MUST): "Stringify returns nil, not an empty map, when the input map is nil or has length zero."
- added `r.every-entry-stringified` (MUST): "Stringify returns a map holding the same keys as the input, where each value is the result of StringifyValue applied to the input value."
- added `r.marshal-failure-falls-back` (MUST): "StringifyValue returns fmt.Sprint of the value when json.Marshal fails, so the function never returns an error and never panics on an unencodable value."
- added `r.non-strings-json-encoded` (MUST): "StringifyValue returns the JSON encoding of a value that is not a string, so numbers become bare numerals, booleans become true or false, and lists become compact JSON arrays."
- added `r.strings-pass-through` (MUST): "StringifyValue returns a value whose dynamic type is string unchanged, without quoting or escaping it."

### common-ref

- intent: "" -> "This context exists so that every layer of kcp-libs agrees on the identity and the transport address of a KCP object. A Ref is the identity used as a map key, a cache index key and a workqueue key, so its string form must be stable and shared; BaseHost and ClusterURL exist so that callers holding either a bare host or a host that already carries a /clusters/<logicalCluster> path can still build a correct per-cluster URL. Keeping these helpers in one tiny package prevents the key format and the KCP API path prefix from being re-implemented, and diverging, in each consumer."
- added `r.base-host-normalization` (MUST): "BaseHost must remove a single trailing slash, and must truncate a host at the first occurrence of the /clusters/ API path prefix, so a host already scoped to a logical cluster reduces to the bare scheme and authority."
- added `r.cluster-url-build` (MUST): "ClusterURL must build the per-logical-cluster URL by normalizing the host with BaseHost and then appending the /clusters/ prefix and the logical cluster, so an already-scoped host does not accumulate a second cluster path."
- added `r.key-format` (MUST): "The key of a reference must be the logical cluster, namespace and name joined by single slash characters, with no other separator or escaping, and both the free function and the method must produce exactly the same string."
- added `r.new-leaves-resource-version-empty` (MUST): "New must construct a Ref from the three identity strings and must leave the resource version empty; a resource version is only ever set through the wither."
- added `r.ref-carries-identity` (MUST): "A Ref must carry the logical cluster, the namespace, the object name and an optional resource version as its four exported fields, so consumers can read and compare identity without an accessor."
- added `r.stdlib-only` (SHOULD): "The package should depend only on the standard library and must not import any kcp client, scheme or workload-runtime package, so every layer of the repository can depend on it without a cycle."
- added `r.with-namespace-returns-copy` (SHOULD): "WithNamespace should return a Ref that carries the replacement namespace and otherwise keeps the receiver's fields unchanged, and should not mutate the receiver."
- added `r.with-resource-version-returns-copy` (MUST): "WithResourceVersion must return a Ref that carries the given version and otherwise keeps the receiver's identity unchanged, and must not mutate the receiver."

### common-statuspatch

- intent: "" -> "This context exists to specify the reusable patch-construction vocabulary shared by every reconcile and store layer in kcp-libs. It keeps the generic JSON-patch byte shapes in one place so controllers do not hand-roll status envelopes, resourceVersion stamps, or finalizer patches, and so the exact wire shape of those patches is pinned by unit tests rather than duplicated per consumer."
- added `r.finalizer-add-single-op` (MUST): "FinalizerAdd MUST encode a single-op JSON patch with op "add" at path /metadata/finalizers and the given finalizers as value, and MUST return a wrapped error prefixed "statuspatch: encode the finalizer patch:" on failure."
- added `r.finalizer-remove-absent-no-patch` (MUST): "FinalizerRemove MUST return a nil body and nil error when the dropped finalizer is not present in current, so that no patch is sent."
- added `r.finalizer-remove-test-and-replace` (MUST): "FinalizerRemove MUST, when the finalizer is present, encode a two-operation JSON patch: first a "test" op on /metadata/finalizers with the original current list as value, then an "add" op on /metadata/finalizers with the remaining finalizers in their original order."
- added `r.finalizers-of-reads-metadata` (MUST): "FinalizersOf MUST decode the body's "metadata" member into Metadata and return its Finalizers list, returning a wrapped error prefixed "statuspatch: decode object metadata:" when decoding fails."
- added `r.merge-wraps-status` (MUST): "Merge MUST encode the given status map as a JSON object under the top-level key "status", returning the encoded bytes, and MUST return a wrapped error prefixed "statuspatch: encode status:" if encoding fails."
- added `r.metadata-decodes-finalizers` (MUST): "Metadata MUST be an exported struct whose Finalizers field is a []string decoded from the JSON key "finalizers"."
- added `r.no-external-dependencies` (SHOULD): "The package SHOULD depend only on the standard library (encoding/json and fmt), keeping the generic patch layer usable without any workload runtime."
- added `r.resourceversion-decode-errors` (MUST): "WithResourceVersion MUST return a wrapped error prefixed "statuspatch: decode patch to stamp a resource version:" when the body is not a JSON object, and "statuspatch: encode patch with a resource version:" when re-encoding fails."
- added `r.resourceversion-empty-passthrough` (MUST): "WithResourceVersion MUST return the input body unchanged when version is the empty string, without decoding or re-encoding it."
- added `r.resourceversion-stamps-metadata` (MUST): "WithResourceVersion MUST decode the body as a JSON object and replace its "metadata" member with an object carrying exactly the resourceVersion key set to the given version, then re-encode it."
- added `r.tests-pin-wire-shape` (SHOULD): "The package's behaviour SHOULD be pinned by direct unit tests covering the status envelope, the resourceVersion stamp including the empty-version case, the test-and-replace finalizer patch, the absent-finalizer nil patch, the single add op, and reading finalizers from an object body."

### common-ttl

- intent: "" -> "This context exists so callers do not each re-derive TTL semantics from raw pointers. Kubernetes-style APIs express TTLs as optional seconds on a timestamp, where nil means unset and negative is meaningless; these helpers centralize that normalization so every consumer agrees on when something is scheduled for deletion, and when nothing should be scheduled at all."
- added `r.deadline-unknown-without-start` (MUST): "Deadline returns false when start is nil, seconds is nil, or seconds is negative. Otherwise it returns true only when now is not before start plus seconds seconds."
- added `r.effective-prefers-override` (MUST): "Effective returns the value of override when override is not nil, otherwise the value of fallback. When the selected value is nil or negative, Effective returns nil. When it is non-negative, Effective returns a pointer to a copy of the value, never an alias of the input."
- added `r.expired-reports-unknown` (MUST): "Expired returns (false, 0, false) when no expiry is known, that is when Expiry reports false. When the expiry is known and now is not before it, Expired returns (true, 0, true). When the expiry is in the future, Expired returns (false, expiry.Sub(now), true)."
- added `r.expiry-needs-completion-and-seconds` (MUST): "Expiry returns the zero time and false when completion is nil, seconds is nil, or seconds is negative. Otherwise it returns completion plus seconds seconds and true."
- added `r.negative-seconds-mean-unset` (MUST): "A negative seconds value is treated exactly like an unset one across Effective, Expiry and Deadline: no expiry and no deadline are produced."
- added `r.tests-cover-boundaries` (SHOULD): "The tests cover override preference and negative rejection in Effective, a future expiry, a past expiry and a missing completion time in Expired, and a future deadline, a reached deadline and a missing start time in Deadline."

### examples-admission

- intent: "" -> "This context exists so that the admission example can be described and regenerated as a compile-checked demonstration of the admission Source interface. It documents the exact contract an implementer must satisfy: how a parent is discovered from a run (a label lookup that may legitimately find nothing), how child runs are enumerated and filtered by that same label, and how capacity is derived from the parent batch, including the not-found case which must surface as a Blocker rather than an error. Keeping the spec anchored to these methods lets changes to the queue capacity and blocker types be validated against a real consumer."
- added `r.capacity-from-batch-spec` (MUST): "source.Capacity must read the parent batch object and return a queue.Capacity carrying the batch spec concurrency policy and max concurrent value."
- added `r.capacity-missing-batch-is-blocker` (MUST): "source.Capacity must return a queue.Blocker with reason BatchMissing and message "the parent batch does not exist" when the batch lookup reports not found via kcpstore.IsNotFound, and must return any other lookup error unchanged instead."
- added `r.example-implements-admission-source` (MUST): "The example source type must satisfy the factory/admission Source interface, whose Parent, Runs and Capacity methods match the example's signatures."
- added `r.parent-lookup-by-label` (MUST): "source.Parent must fetch the run object by its ref, read the parent label from the object metadata, and return false with a zero ref when the label is empty, so a run without a parent is not an error."
- added `r.parent-ref-preserves-cluster-and-namespace` (MUST): "source.Parent must build the parent ref from the run's own logical cluster and namespace, using the label value as the name."
- added `r.run-entrypoint-writes-to-writer` (MUST): "Run must accept a context and an io.Writer and return an error, using the writer for the example's output so the example can be driven from a test."
- added `r.runs-filter-by-parent-label` (MUST): "source.Runs must skip objects whose parent label does not equal the parent name, and emit one queue.Run per remaining object carrying its ref, status phase and creation timestamp."
- added `r.runs-lists-within-cluster` (MUST): "source.Runs must list items in the parent's logical cluster and return the error unchanged when the list fails."
- added `r.runs-parse-creation-timestamp` (MUST): "source.Runs must parse the object creation timestamp with time.RFC3339Nano and leave the zero time when parsing fails rather than returning an error."

### examples-controller

- intent: "" -> "This context exists to demonstrate, end to end and against a real kcp instance, that the generic controller layer works: a store backed by kcp, an export watch that discovers the virtual workspace URL, an informer-based controller with indexers and a bounded reconcile policy, and metrics. It is an executable example and a test fixture (main_test.go), not library code, so it composes the generic packages (kcpstore, exportwatch, cache, informerwatch, controller, metrics) rather than defining new abstractions."
- added `r-await-exported-workspace-url` (MUST): "Run must build a kcpstore, await the exportwatch endpoints for livekcp.Export, derive the base paths, and return an error when kcp published no virtual workspace URL for the export."
- added `r-report-controller-state` (MUST): "Run must report the reconcile count, the queue depth, the cache age and the rendered metric registry to the output writer before returning."
- added `r-seed-and-wait-for-widgets` (MUST): "Run must seed the widgets alpha and beta, start the controller, and wait for each of them to reach the succeeded phase before seeding gamma and waiting for it likewise."
- added `r-test-covers-run` (SHOULD): "The controller example should be covered by a test that drives Run and checks its output."
- added `r.run-is-the-entry-point` (MUST): "The example must expose Run(ctx context.Context, out io.Writer) error as its entry point, returning an error and writing all human-readable output to out."
- added `r.start-and-stop-cluster` (MUST): "Run must start a live kcp cluster and stop it on return, and must return the error if the cluster fails to start."

### examples-dns

- intent: "" -> "This context exists to demonstrate and verify the DNS and service-name resolution behaviour of kcp-libs end to end. It is an executable example rather than library code: it shows a consumer of `servicenames` how to assemble a resolver over several logical clusters, how tokens for the kcp server are minted and injected into a pod environment, and how the DNS shim and readiness probe assets are placed on disk so a workload can resolve sibling services. The test in `main_test.go` pins the observable output of that flow, so a change in table contents, token minting, asset placement, advertised address, or label rendering fails the example."
- added `r.advertised-address` (SHOULD): "A service that binds to the wildcard host must advertise a reachable loopback address derived from its environment block."
- added `r.asset-materialisation` (MUST): "`Run` must materialise the DNS shim and probe assets into a temporary directory, use the shim path as the resolver shim, and use `kcpstore.NewPathCache` as the resolver path cache."
- added `r.label-rendering` (SHOULD): "The example must render the workspace path in its label form through the kcp service labels helper."
- added `r.output-pinned-by-test` (MUST): "The example output must contain the exact substrings asserted by `TestRunBuildsTheTableAndTheTokens`: the three name resolutions, the workspace count, the injected name and token counts, the self name resolution, the minted token line, the asset directory and file names, the preload line, the probe grant list, the advertised address, and the label rendering."
- added `r.probe-grants` (MUST): "The readiness probe must preload the same shim file as the workload and be granted the DNS probe environment variables."
- added `r.run-entry-point` (MUST): "The example package must expose `Run(ctx context.Context, out io.Writer) error` and write all of its reporting to `out` rather than to the process standard output."
- added `r.seeded-workspaces` (MUST): "`Run` must seed a pod named `pds` and a pod named `api` into the first consumer cluster and a pod named `pds` into the second consumer cluster, each with an environment map that names a port or an args list."
- added `r.self-name-first-moment` (MUST): "A pod's own name must already appear in the injected table at the moment its environment is built, resolving to the address of its own deployment."
- added `r.table-across-workspaces` (MUST): "The resolver source must list pods from both consumer workspaces and annotate each object with its logical cluster, so the resulting table covers both workspaces."
- added `r.test-requires-live-cluster` (MUST): "The test must call the live cluster requirement guard before running the example so it skips when no live kcp environment is available, and must bound the run with a five minute context."
- added `r.token-minting` (MUST): "The resolver must mint tokens through the kcp store with a one hour TTL, and its environment output must carry a token for each covered workspace."

### examples-pki

- intent: "" -> "This context exists to document the examples/pki example: an executable end-to-end demonstration that provisions a PKI setup through KCP and issues certificates, backed by an in-process fake Vault. The fake Vault exists so the example and its test can run without a real Vault server, while still recording every HTTP call for assertions."
- added `r.cert-helpers` (SHOULD): "Certificate handling for the example SHOULD live in certs.go and stay separate from the entry point and the fake vault."
- added `r.pki-tested` (SHOULD): "The example SHOULD stay covered by main_test.go, which runs it against the fake vault."
- added `r.run-entry-point` (MUST): "The example MUST expose Run, which takes a context and an output writer and returns an error, as its executable entry point."
- added `r.vault-call-count` (MUST): "The fake vault MUST expose CallCount, returning the number of calls it has recorded, so a caller can assert on observed traffic."
- added `r.vault-close` (MUST): "The fake vault MUST expose Close, returning an error, so its HTTP server is shut down after use."
- added `r.vault-serves-vault-api` (MUST): "The fake vault MUST implement ServeHTTP over http.ResponseWriter and *http.Request so it can act as an HTTP handler for the Vault API paths the example calls."
- added `r.vault-url` (MUST): "The fake vault MUST expose URL, returning the base URL the example uses to reach it."

### examples-policy

- intent: "" -> "This context documents the examples/policy package: the runnable demonstration of the policy engine surface. It exists so the exported behaviour of that example — the engine's URL accessor, its Close, its submitted-run counter, its HTTP handler, and the Run entry point — is described and anchored to code, letting a reader or a tool reason about what the example guarantees without reading every line."
- added `r.engine-close` (MUST): "engine.Close must shut down the embedded HTTP server and must return the error from that server's Close."
- added `r.engine-http-handler` (MUST): "engine.ServeHTTP must take an http.ResponseWriter and an *http.Request, so the engine satisfies http.Handler and can be mounted on a net/http server."
- added `r.engine-submit-count` (MUST): "engine.Submits must return the recorded submit count, read while holding the engine mutex, so concurrent callers see a consistent value."
- added `r.engine-url` (MUST): "engine.URL must return the address the engine's own HTTP server listens on, so a caller can hand the endpoint to the policy client."
- added `r.example-tested` (MAY): "The example may carry a test file, main_test.go, exercising the Run entry point and the engine."
- added `r.run-entry-point` (MUST): "Run must be the example's entry point, taking a context.Context and an io.Writer and returning an error; it must use the engine's URL, Close and Submits methods to drive and verify the run."

### examples-workloads

- intent: "" -> "This context exists to prove the workload packages compose in one place and to serve as living documentation for callers. It is an example, not a library: it is a package main whose only exported symbol is Run, deliberately taking a context and an io.Writer so the same code path can be driven by a test that captures output instead of the terminal. The accompanying test asserts on the printed lines, which makes the example double as an integration check that the deno permission vocabulary, the exec runner's argv layout, probe thresholds, ttl arithmetic, job allocation bookkeeping and the memory runner all keep behaving as documented."
- added `r.exec-pod-lifecycle` (MUST): "Run starts a workload through the execrunner pod, asserts it satisfies runner.PodRunner, waits until the run leaves the running state, and reports the terminal state and the answer output."
- added `r.failure-and-probe-paths-shown` (MUST): "Run must also exercise a failing run that reports a non-zero exit code, a readiness probe that passes, and a liveness tracker that requests a restart once the failure threshold is reached."
- added `r.memory-runner-parity` (SHOULD): "Run replays the greeting script through the in-memory runner and shows it produces the same answer without starting a process."
- added `r.permissions-become-argv` (MUST): "Run converts a denospec.Permissions value with a net allow-list and denied env into argument strings via denospec.Args, and prints them."
- added `r.run-is-the-entry-point` (MUST): "The example exposes Run(ctx context.Context, out io.Writer) error; it takes the context and the output writer as parameters and returns an error rather than terminating the process, so a caller other than main can drive it."
- added `r.self-cleaning-temp-dir` (MUST): "Run creates a temporary directory and removes it on return, and places both the materialised runtime binary and the runs directory underneath it."
- added `r.test-pins-printed-output` (MUST): "The test drives Run with a captured buffer and fails unless the output contains the expected runtime, permissions, argv, terminal states, probe, liveness, allocation and in-memory-runner lines."
- added `r.ttl-and-joballoc-shown` (SHOULD): "Run evaluates ttl.Expired for a past completion time and uses joballoc to allocate a job's names and report the still-pending subset."

### factory-admission

- intent: "" -> "This context exists to describe the admission gate that sits between a queue and the workloads it schedules. It defines the Source port through which the gate reads cluster state (parent, siblings, capacity) without knowing any workload runtime, and the Admitter that turns those reads plus a lease table into a queue.Admission decision and into wake notifications for runs that have become eligible. It is the piece that enforces per-parent concurrency policy, so callers can ask whether a specific run may proceed and can be told which runs to wake once capacity frees up."
- added `r.admit-grant` (MUST): "Admit must grant a lease for the run on the parent at the current time only when the resulting admission is both Gated and Allowed."
- added `r.admit-no-parent` (MUST): "Admit must return an empty queue.Admission and a nil error when Source.Parent reports the run has no parent, and must propagate any error from Parent, Runs or Capacity without deciding."
- added `r.admit-plan` (MUST): "Admit must build the observed phase map from the sibling runs, count reserved leases for the parent at the current time, add the candidate run to the sibling list only when it is absent, and return the queue.PlanIndex entry for that run's ref."
- added `r.leases-accessor` (MUST): "Admitter.Leases must return the lease table held in the options."
- added `r.new-defaults` (MUST): "New must default Leases to queue.NewLeases(LeaseTTL) when nil and Now to time.Now when nil, and return an Admitter holding the resulting Options."
- added `r.options-shape` (MUST): "Options must carry Source, Leases, LeaseTTL, Now, Lifecycle, Wake, RunKind and ParentKind, and Admitter must hold those options unexported."
- added `r.source-capacity` (MUST): "Source.Capacity must report the parent's capacity together with the blocker that limits it, so the plan can be computed and the limit derived."
- added `r.source-parent` (MUST): "Source.Parent must report the parent of a run, with a boolean that is false when the run has no parent and an error that aborts the decision."
- added `r.source-port` (MUST): "Admission state must be read through the Source interface only, which exposes Parent, Runs and Capacity and names no workload runtime."
- added `r.source-runs` (MUST): "Source.Runs must list the runs already present under a parent, which the admitter uses to build the observed phase map and to size wake lists."
- added `r.test-coverage` (SHOULD): "The package should be exercised through a stub Source in tests, as factory/admission/admission_test.go does."
- added `r.wake-noop` (MUST): "Wake must return nil without touching the Source when no Wake callback is configured."
- added `r.wake-parent` (SHOULD): "Wake should fire the Wake callback with ParentKind and the parent ref after the per-run wakes, and only when ParentKind is non-empty."
- added `r.wake-runs` (MUST): "Wake must re-read the parent's runs and capacity, derive the limit from the capacity policy and max concurrency, and fire the Wake callback with RunKind for each run returned by queue.WakeList for the free slot count, where free is the limit minus the running count and is clamped at zero."

### factory-controller

- intent: "" -> "This context is the controller runtime of kcp-libs: the generic work-queue driver that turns watched kcp resources into reconcile invocations. It exists so the deno workload packages and downstream consumers can share one tested scheduling core — bounded-concurrency workers, rate limited retries, optimistic-concurrency conflict handling, informer-event freshness tracking and prometheus metrics — instead of each rebuilding it. It names no workload runtime; the deno vocabulary lives in the consumers, and the controller only knows kinds, refs and a reconcile.Handler."
- added `r.default-workers-clamped` (MUST): "DefaultWorkers returns runtime.GOMAXPROCS(0) clamped to the range 1 through 16."
- added `r.enqueue-key` (MUST): "Enqueue builds a reconcile.Key from the kind and a ref.New of the ref.Ref's LogicalCluster, Namespace and Name, and adds it to the rate limiting workqueue."
- added `r.event-freshness` (SHOULD): "The OnEvent hook wired into informerwatch records the current Now timestamp so that CacheAge and cache_age_seconds reflect informer liveness."
- added `r.live-verification` (SHOULD): "The controller is exercised against a real kcp cluster by TestLiveControllerAgainstRealKCP using kcpstore and exportwatch wiring."
- added `r.metrics-families` (MUST): "New registers reconcile_seconds (summary), conflicts_total (counter), queue_depth (gauge function over the queue length) and cache_age_seconds (gauge function over CacheAge, reporting NaN before the first event) on the chosen metrics registry."
- added `r.new-defaults-options` (MUST): "New fills absent options: Policy.Interval from reconcile.DefaultRequeueAfter, Policy.MinTransitionPoll from reconcile.DefaultMinTransitionPoll, Workers from DefaultWorkers, Now from time.Now, Log from slog.Default, IsConflict from apierrors.IsConflict, Metrics from metrics.New("controller") and Set from cache.NewSet."
- added `r.new-validates-wiring` (MUST): "New returns an error when Config is nil ("controller: a rest config is required"), when Handler is nil ("controller: a handler is required"), or when Sources is empty ("controller: at least one source is required"), and returns a usable *Controller otherwise."
- added `r.observability-readers` (MUST): "QueueDepth returns the workqueue length, Reconciles, Conflicts and Errors return the respective atomic counters, and CacheAge returns zero before the first informer event and otherwise the elapsed time from the last event to Now."
- added `r.run-supervises` (MUST): "Run starts informerwatch.Run in a goroutine with Config, Sources, Indexers, Set, Reactor, Enqueue and OnEvent wired from the options, starts one worker per Workers, then waits on ctx.Done() or the watch error channel; it shuts the queue down, waits for all workers, and returns the watch error unless it is context.Canceled."
- added `r.worker-conflict-path` (MUST): "A worker that receives an error matching opts.IsConflict increments the conflicts counter and the conflicts_total metric, forgets the key, re-queues it after Policy.ConflictAfter(), and does not increment the errors counter or log."
- added `r.worker-error-path` (MUST): "A worker that receives a non-conflict error increments the errors counter, logs "reconcile failed" with the kind, workspace, name and error, and re-queues the key rate limited; the handler call is timed into the reconcile_seconds summary."
- added `r.worker-success-path` (MUST): "A worker that handles a key without error forgets it, asks Policy.Next(key, after, terminal) for a delay and requeue decision, and re-queues the key after that delay only when requeue is true."

### factory-servicenames

- intent: "" -> "This context exists so that workloads running under kcp can discover each other by name without a hand-written DNS configuration. It converts the pods a cluster controller already observes into a service table (name to advertised host:port), the workspace list worth addressing, and the service account tokens needed to reach them, then hands that whole bundle to a workload as environment variables. It also defines the seams (Source, PathResolver, TokenMinter, the workspace source) that keep the package free of any particular client, so a caller supplies real store-backed pods, paths and workspaces while tests supply plain functions. The token map and the service names agree on one cluster naming on purpose: a workload discovers a peer by parsing the peer's name back to its cluster and looking the token up by it, so a token keyed any other way is a peer it can never reach."
- added `r.advertised-address` (MUST): "AdvertisedAddress parses the args JSON as a string list and the env JSON as a string map, prefers the --port and --hostname flag values over PORT and HOSTNAME, returns the empty string when no port is found, and substitutes 127.0.0.1 when the host is empty, "0.0.0.0" or "::"; the result is host:port."
- added `r.behaviour-is-tested` (MUST): "The package behaviour is covered by tests in the same package, including the table path lookup, the self name carried before the pod is observed, the shim preload entry, and per-workspace token minting."
- added `r.env-bundle` (MUST): "Resolver.Env always sets the domain and the target namespace, sets the shim key only when Options.Shim is non-empty, adds the JSON table only when the table is non-empty, adds the tokens key whenever an account is given, and folds the target's own advertised address into the table and the workspace list when that address is non-empty."
- added `r.name-uses-path` (MUST): "Resolver.Name routes the logical cluster through the PathResolver when one is set and it returns a non-empty path, and otherwise uses the logical cluster unchanged, then returns kcp.ServiceFQDN of name, namespace and cluster under the configured domain."
- added `r.new-defaults` (MUST): "New fills Options.Domain with kcp.DefaultServiceDomain and Options.ServiceAccountNamespace with "default" when either is empty, and returns a Resolver holding the resulting options."
- added `r.path-resolver-lookup` (MUST): "A PathResolver maps a logical cluster to a path, returning the empty string when it has no answer, and PathResolverFunc adapts a function with the same signature."
- added `r.source-provides-pods` (MUST): "A Source returns the observed pods as a slice of unstructured objects, and SourceFunc adapts a plain function to that interface by calling it."
- added `r.table-skips-unlabelled-and-unadvertised` (MUST): "Resolver.Table returns an empty table and no workspaces when no Source is set; it skips pods lacking the kcp.ClusterAnnotation, records each distinct cluster once in the returned workspace list in first-seen order, skips pods whose advertised address is empty, and maps each remaining pod's resolved name to its advertised address."
- added `r.token-keys-agree-with-the-service-name` (MUST): "Every key in the token JSON is the same cluster string Resolver.Name builds the service name from for that cluster -- the path the PathResolver answers when it answers a non-empty path, and the cluster unchanged otherwise. A workload parses a service name back to that cluster (pds.default.alice.svc.kcp.local to root:alice) and looks its token up by it, so a key that is an unresolved cluster id, or a cluster named any other way, is a lookup the workload can never hit and a peer it can never discover."
- added `r.tokens-per-workspace` (MUST): "Resolver.Tokens mints one service account token per workspace against the account's namespace, falling back to Options.ServiceAccountNamespace when the account namespace is empty, skips workspaces whose mint fails, and returns the result as a JSON object string, or "{}" when marshalling fails. It mints nothing when the account is nil or no Minter is set."
- added `r.workspace-source-tests` (MUST): "Tests cover it. Naming a cluster that has no pod still mints a token under the key the service name for that cluster implies, the target's own cluster is always in the set, a nil source adds nothing, and a cluster named twice is minted once. Every file the change touches is gofmt-clean, as every change to this repository must be."
- added `r.workspace-source-widens-the-token-set` (MUST): "Options carries a workspace source -- an interface whose Clusters method returns the logical clusters a workload may address, and no clusters when it has no answer. The set of workspaces a token is minted for is the union of the clusters the table names, the target's own cluster, and the clusters that source returns, deduplicated. A cluster with no pod yet is therefore still given a token, which is what lets a workload discover a service created after it started by name; a nil source leaves the set exactly as the table and the target make it."

### impl-assets

- intent: "" -> "This context exists so callers can hand a declarative set of asset files to one object and get back a stable name-to-path map, without caring when or how often the bytes hit disk. It separates one-shot materialisation, which is idempotent and retried after failure, from explicit rewriting, which always writes, and it gives the DNS workload a ready-made set for its shim and probe scripts."
- added `r.dns-assets-embedded` (MUST): "The shim and probe scripts shipped for the DNS workload live in the package as dnsshim.ts and dnsprobe.ts and are the bytes written by DNSSet."
- added `r.dnsset-builds-run-dir-set` (MUST): "DNSSet builds a Set whose Dir is the DNS directory under the given runs directory and whose Files carry the shim source under ShimName and the probe source under ProbeName."
- added `r.dnsshim-discovers-a-name-absent-from-the-table-test` (MUST): "A deno test covers discovery, in the way the shim's other behaviour is covered. With a table that does not name the target, a token map that carries the target's cluster, and a local API server answering the target's denopod object, a fetch of the target's service name reaches the object's advertised address. With the same server and table but a token map keyed by the cluster id instead, the fetch reports that the name could not be discovered -- the key mismatch this guards against -- and no request is sent."
- added `r.dnsshim-fetch-preserves-the-request` (MUST): "The shim replaces globalThis.fetch, and when the name it is given is in the table it must forward the caller's request unchanged apart from the connection target. A caller that passes a Request object must reach the real fetch as that same request aimed at the table address: its method, body, headers, redirect mode, credentials, signal and cache mode are preserved. It must not be forwarded as a bare URL with an empty init, which drops the Request and turns a POST into a GET -- a service that routes by method then answers 404 for a write that was meant to create a resource, which is what a DID registration did. A caller that passes a URL and an init keeps working exactly as before, the Host header the shim adds carries the original name, and a name outside the table is still passed through untouched."
- added `r.dnsshim-fetch-preserves-the-request-test` (MUST): "A test covers it, in the way the shim's other behaviour is covered -- deno runs the shim with the table pointed at a local listener, a driver calls fetch with a Request object whose method is not GET and whose body is known, and the listener records what it received. The test asserts the listener saw that method and that body, and it must fail against the shim before this requirement."
- added `r.dnsshim-token-key-matches-the-name` (MUST): "When a name is absent from the table the shim falls back to discovery, and it reads the token for the target from the injected token map under the cluster it derives from the service name itself (pds.default.alice.svc.kcp.local to root:alice) -- the same cluster the provider keyed that token by. It must not read it under the raw cluster id or any other key, and when the map holds no token for that cluster it gives up instead of sending an unauthenticated request."
- added `r.materialise-failure-retried` (MUST): "A failed Materialise leaves the set unwritten and stores the error, so a later Materialise attempt writes again instead of returning the cached failure."
- added `r.materialise-writes-once` (MUST): "Set.Materialise writes the files only when the set has not already been written, records success in written, and returns the published name-to-path map with the recorded error."
- added `r.path-resolves-one-file` (MUST): "Set.Path returns the on-disk path of the named asset, materialising the set first, and reports an error for a name the set does not carry."
- added `r.rewrite-always-writes` (MUST): "Set.Rewrite writes the files on every call, ignoring the written flag, and marks the set written on success or unwritten on failure."
- added `r.set-carries-dir-files-perm` (MUST): "Set exposes Dir, Files (name to bytes), and Perm so a caller can declare where assets go, what they contain, and what mode they get."
- added `r.set-is-concurrency-safe` (MUST): "Set guards its written flag, path cache, and error with a mutex so concurrent Materialise, Rewrite, and Path calls stay consistent."

### impl-execrunner

- intent: "" -> "This context exists so the execrunner package has a written specification of the process-execution engine and pod: what the two option structs and their constructors require and default, and what the four lifecycle operations on each runner (Start, Observe, Stop, Probe) must do. It anchors the contract that the runner types implement the generic runner interfaces without knowing about Deno specifics beyond launching a binary, and it records the invariants that the tests in pod_test.go pin down: required directories, path resolution, timeout handling, and non-interference with caller-supplied environment."
- added `r.engine-denobin-default` (MUST): "NewEngine defaults EngineOptions.DenoBin to "deno" when the caller leaves it empty."
- added `r.engine-paths-absolute` (MUST): "NewEngine resolves ServerDir and RunsDir to absolute paths before storing them in the engine options."
- added `r.engine-runs-server` (MUST): "Engine.Start must run the server file inside the server directory under the runs directory and return a run identifier that Engine.Observe, Engine.Stop and Engine.Probe accept."
- added `r.engine-runsdir-required` (MUST): "NewEngine returns an error when EngineOptions.RunsDir is empty."
- added `r.engine-serverdir-required` (MUST): "NewEngine returns an error when EngineOptions.ServerDir is empty."
- added `r.engine-serverfile-default` (MUST): "NewEngine defaults EngineOptions.ServerFile to "main.ts" when the caller leaves it empty."
- added `r.namespace-left-alone` (MUST): "The pod must not set an environment variable for the namespace that overrides the one the caller supplied in the pod request."
- added `r.no-deadline-never-reaped` (MUST): "A run started without a deadline must never be reaped by the timeout machinery."
- added `r.pod-result-and-tls-options` (MAY): "PodOptions may carry a result file path, CA data and a trust bundle function for pods that need to talk to TLS endpoints or publish a result."
- added `r.pod-runs-script` (MUST): "Pod.Start must run the request script in the pod's run directory and return a run identifier that Pod.Observe, Pod.Stop and Pod.Probe accept."
- added `r.pod-runsdir-required` (MUST): "NewPod returns an error when PodOptions.RunsDir is empty, since the pod has no run directory to work in."
- added `r.probe-runs-command` (SHOULD): "Probe should run the supplied command against the run under the supplied timeout and report whether it succeeded."
- added `r.shared-supervisor` (SHOULD): "Engine and Pod should share the supervisor in impl/execrunner/supervisor.go for run directories, process lifetime and timeouts rather than each managing processes themselves."
- added `r.stop-reports-wait-error` (MUST): "Engine.Stop must report the wait error observed when the process is stopped."
- added `r.timeout-enforced` (MUST): "When a timeout is configured, the engine must enforce it against the running process."

### impl-exportwatch

- intent: "" -> "This context exists so callers can wait until a provider's exports are actually reachable before proceeding: it isolates the endpoint-slice discovery, projection and readiness logic from any workload runtime, gives a single place where the provider workspace is appended to the API host path, and exposes both a blocking wait (Await) and a watch-driven wait (WatchEndpointSlices) over the same Options."
- added `r.await-returns-discovered-endpoints` (MUST): "Await takes the same options as Discover and returns the discovered Endpoints once they are ready, so a blocking wait and a one-shot discovery share one configuration."
- added `r.client-requires-config` (MUST): "Client refuses a nil rest config with the error "exportwatch: a rest config is required", so a client is never built without one."
- added `r.client-scopes-host-to-workspace` (MUST): "Client applies the client limit tuning to the config and rewrites the host to the options host plus the API path prefix plus the provider workspace, so every request is scoped to that workspace."
- added `r.discover-lists-endpoint-slices` (MUST): "Discover builds a client from the options, lists the endpoint slice resource, and wraps a list failure as "exportwatch: list the endpoint slices: %w"; on success it returns FromList of the result."
- added `r.endpoints-map-shape` (MUST): "Endpoint addresses are held as a map from export name to a slice of URL strings, so one export can carry several addresses."
- added `r.fromlist-reads-export-and-urls` (MUST): "FromList reads spec.export.name from each item and the url of each entry in status.endpoints, skipping items with no export name and entries with no url, and appends the urls under that export name."
- added `r.options-carries-host-resolution` (MAY): "Options exposes an unexported host method that yields the API host, so the host derivation stays with the options that describe the provider workspace."
- added `r.paths-renders-one-export` (MUST): "Paths takes the discovered Endpoints and a single export name and returns that export's addresses as a string slice."
- added `r.ready-requires-every-export` (MUST): "Endpoints.Ready reports false when any named export has zero addresses and true only when every named export has at least one."
- added `r.watch-bounds-itself` (MUST): "WatchEndpointSlices derives a context bounded by the wait duration, stops the watch on return, and reports success (nil) when that context ends or the result channel closes."
- added `r.watch-signals-readiness` (MUST): "On an Added, Modified or Deleted event WatchEndpointSlices rediscovers the endpoints and returns ErrReady as soon as the options exports are ready; a failed rediscovery is ignored rather than fatal."

### impl-informerwatch

- intent: "" -> "This context exists so the kcp-libs dynamic watching layer is described as it is written: a runtime-neutral informer bootstrap that turns kcp object events into enqueue calls and optional reactor callbacks, with the indexer, cache set and reactor hooks injected by the caller. It names no workload runtime and is reusable without the deno packages."
- added `r.block-until-done` (MUST): "After all caches sync, Run must block on the context's Done channel and return nil when the context is cancelled."
- added `r.client-per-source` (MUST): "Run must build one dynamic client per Source, with the client host set to the source Base (trailing slash trimmed) plus ref.APIPathPrefix and a wildcard, and must apply common/clientlimit to the rest config."
- added `r.enqueue-own-object` (MUST): "Add and update events must enqueue the changed object itself: the handler resolves the object with cache.RefOf and calls Enqueue with the resource Kind and the resolved ref, doing nothing when the object yields no ref."
- added `r.indexers-applied` (MUST): "Registration must add Options.Indexers to the informer indexer when they are non-nil, and must fail with an error naming the Kind when the indexer cannot be added."
- added `r.kind-watched-once` (MUST): "Run must refuse a Kind that appears in more than one Resource, because the second watcher would replace the first one's index; the error names the Kind and says it is watched twice."
- added `r.on-event-hook` (SHOULD): "The add and update handlers should call Options.OnEvent when it is non-nil, before any enqueue or reactor work."
- added `r.reactor-added` (MUST): "On an add event with Options.Reactor set, the handler must call Reactor.Added with the Kind, the object as *unstructured.Unstructured, and the enqueue function, but only when the object is an *unstructured.Unstructured."
- added `r.reactor-updated` (MUST): "On an update event with Options.Reactor set, the handler must call Reactor.Updated with the Kind, the old object (as *unstructured.Unstructured when it converts, otherwise nil), the new object as *unstructured.Unstructured, and the enqueue function."
- added `r.register-per-resource` (MUST): "Run must register an event handler for every Resource of every Source on a filtered dynamic shared informer factory started with the context's Done channel."
- added `r.run-requires-enqueue` (MUST): "Run must return the error "informerwatch: an enqueue function is required" when Options.Enqueue is nil."
- added `r.run-requires-rest-config` (MUST): "Run must return the error "informerwatch: a rest config is required" when Options.Config is nil."
- added `r.set-populated` (SHOULD): "Registration should add the informer indexer to Options.Set under the resource Kind when Set is non-nil."
- added `r.source-needs-base` (MUST): "Run must reject a Source whose Base is empty with the error "informerwatch: every source needs a base URL"."
- added `r.source-needs-resource` (MUST): "Run must reject a Source with no Resources with the error "informerwatch: every source needs at least one resource"."
- added `r.wait-for-sync` (MUST): "Run must wait for every started factory's cache to sync; when a cache fails to sync it must return nil if the context is already done, and otherwise the error "informerwatch: the cache for %s did not sync" naming the resource."

### impl-kcpstore

- intent: "" -> "This context exists so that consumers (controllers, admission, DNS, servicenames) have one place that knows how to reach kcp: URL shape, per-cluster client reuse, rate limiting, JSON decoding, status and finalizer patching, token minting, and workspace path resolution. The path cache is the newest part and it is deliberately conservative — it exists because a workspace path that failed to resolve must not be frozen into a wrong name, so only an answer that cannot change by asking again is stored."
- added `r.add-finalizer` (MUST): "Resource.AddFinalizer must build the addition patch with statuspatch.FinalizerAdd and apply it through PatchJSON."
- added `r.cache-mutex` (MUST): "PathCache and Store must guard their maps with a mutex, and neither Lookup nor ClientFor may hold the lock across the network call."
- added `r.client-per-cluster-gv` (MUST): "ClientFor must key clients on logicalCluster plus the group-version string, return a cached client when one exists, and otherwise build one from a copy of the store config with APIPath set to ref.APIPathPrefix + logicalCluster + "/apis", or + "/api" when the group is empty."
- added `r.cluster-path-annotation` (MUST): "Store.ClusterPath must read the logicalclusters/cluster object in core.kcp.io/v1alpha1 and return its kcp.PathAnnotation value, failing with ErrNoPath when the annotation is empty and wrapping read and parse failures with the logical cluster named."
- added `r.config-accessor` (MUST): "Store.Config must return the store's tuned rest.Config."
- added `r.config-tuning` (MUST): "New must derive the store config by applying the QPS and Burst limits to the supplied RestConfig, or to an empty config when RestConfig is nil, then set Host from ref.BaseHost, JSON content type and accept types, and the supplied Transport when non-nil."
- added `r.create-from-metadata` (MUST): "Resource.Create must marshal the object, read its namespace from the encoded metadata, POST it into that namespace, and on failure include the response body detail plus the object name in the wrapped error."
- added `r.delete-background` (MUST): "Resource.Delete must delete with propagationPolicy Background and treat an already-absent object as success, returning nil for IsNotFound errors."
- added `r.error-wrapping-prefix` (SHOULD): "Errors leaving the package should carry the "kcpstore:" prefix and name the resource, object, and logical cluster involved so callers can attribute a failure without re-reading the request."
- added `r.finalizers-read` (MUST): "Resource.Finalizers must read the raw object and return the finalizers parsed by statuspatch.FinalizersOf."
- added `r.generic-resource` (MUST): "Of must return a Resource[T] bound to a Store and a GroupVersionResource, and every Resource method must route through the client for the target's logical cluster."
- added `r.get-decode` (MUST): "Resource.Get must fetch the raw object and unmarshal it into a T, failing with "kcpstore: decode object: <err>" when the body does not parse."
- added `r.get-raw` (MUST): "Resource.GetRaw must GET the named resource in the target namespace and logical cluster and return the raw bytes, wrapping a failure as "kcpstore: read <resource> <name> in <cluster>: <err>"."
- added `r.host-required` (MUST): "New must reject an Options whose Host is empty with the error "kcpstore: Host is required"."
- added `r.list-cluster` (MUST): "Resource.List must GET the resource across the logical cluster and decode the items array, wrapping failure as "kcpstore: list <resource> in <cluster>: <err>"."
- added `r.list-namespace` (MUST): "Resource.ListIn must GET the resource in one namespace of a logical cluster and decode the items array, wrapping failure as "kcpstore: list <resource> in <cluster>/<namespace>: <err>"."
- added `r.live-tests` (MAY): "The package may be exercised against a real kcp instance through the live tests alongside the unit tests."
- added `r.mint-token` (MUST): "Store.MintServiceAccountToken must default an empty namespace to "default", POST a TokenRequest with the TTL converted to seconds at the serviceaccounts/<name>/token subresource in the core v1 group, and fail when the response carries no token."
- added `r.patch-json` (MUST): "Resource.PatchJSON must apply the patch as a JSON patch to the named resource, failing with "kcpstore: patch <resource> <name> in <cluster>: <err><detail>"."
- added `r.patch-status-merge` (MUST): "Resource.PatchStatus must apply the patch at the status subresource as a merge patch after statuspatch.WithResourceVersion stamps the target's resource version, and must fail with the response detail on error."
- added `r.path-cache-definitive-only` (MUST): "PathCache.Lookup must store only what ClusterPath answers definitively: success, ErrNoPath, or an IsNotFound error. Every other error must return the empty string without writing an entry, so it is asked again."
- added `r.path-cache-forget` (MUST): "PathCache.Forget must drop the entry for a logical cluster so the next Lookup resolves it again, and PathCache.Len must report the number of cached entries; both must be safe under the cache mutex."
- added `r.path-cache-hit-skips-store` (MUST): "PathCache.Lookup must return a cached path without calling the store when the logical cluster is already present, and must tolerate a nil store by caching the empty string."
- added `r.path-cache-retry-healing-errors` (MUST): "A refusal, an outage, an unparseable body, and a cancelled call must all be treated as non-definitive, because a token mid-refresh and a grant not yet visible are refusals that heal; the allowlist form means an unanticipated error is retried rather than frozen into a workspace's name."
- added `r.remove-finalizer-absent-ok` (MUST): "Resource.RemoveFinalizer must read the current finalizers and remove one, treating an object that is no longer there as success, and it must delegate the patch to RemoveKnownFinalizer."
- added `r.remove-known-finalizer-noop` (MUST): "Resource.RemoveKnownFinalizer must build the removal patch from the supplied current list and skip the request entirely when statuspatch.FinalizerRemove returns a nil patch."

### impl-memoryrunner

- intent: "" -> "This context exists to give the rest of kcp-libs a controllable, in-memory fake for the runner contracts (abc/runner), so controllers, reconcile loops and tests can drive pod and engine lifecycles through Start/Observe/Stop/Probe without spawning real processes. It carries the observed tests, which pin the observable behavior: a pod runs for PollsBeforeDone polls before reporting the configured outcome, Stop on a running pod yields StateFailed "stopped", an unknown pod run is an error, and an engine exits after ExitsAfterPolls polls or when stopped with message "stopped". It is deliberately minimal and lives in its own package so the generic layers stay free of any workload-runtime assumption."
- added `r.engine-construction` (MUST): "NewEngine must initialise the run map and store the given EngineOptions without applying defaults."
- added `r.engine-exits-after-polls` (MUST): "Engine.Observe must increment the run's poll counter and, when EngineOptions.ExitsAfterPolls is greater than zero and the run is StateRunning with polls greater than or equal to it, replace the status with runner.EngineStatus{State: runner.StateFailed, Message: "exited"}."
- added `r.engine-observe-unknown-run-errors` (MUST): "Engine.Observe must return an error when the run id is unknown, formatted as "memoryrunner: unknown engine run %s"."
- added `r.engine-probe-always-true` (MUST): "Engine.Probe must ignore its arguments and return true with a nil error."
- added `r.engine-satisfies-enginerunner` (MUST): "Engine must satisfy runner.EngineRunner, asserted at compile time with var _ runner.EngineRunner = (*Engine)(nil)."
- added `r.engine-start-running` (MUST): "Engine.Start must allocate an id formatted as mengine-<n> from an atomic sequence and record the run in StateRunning, returning the id and a nil error."
- added `r.engine-stop-marks-failure` (MUST): "Engine.Stop must set a known, still-running run to runner.EngineStatus{State: runner.StateFailed, Message: "stopped"} and return nil in every case."
- added `r.no-runtime-dependency` (SHOULD): "The package should stay free of process spawning, filesystem or network effects and depend only on abc/runner, so it remains a deterministic stand-in for real runners."
- added `r.pod-observe-transitions-on-polls` (MUST): "Pod.Observe must increment the run's poll counter and, while the run is StateRunning and polls are greater than or equal to PodOptions.PollsBeforeDone, replace the status with PodOptions.Outcome; earlier observes report StateRunning."
- added `r.pod-observe-unknown-run-errors` (MUST): "Pod.Observe must return an error when the run id is unknown, formatted as "memoryrunner: unknown pod run %s"."
- added `r.pod-option-defaults` (MUST): "NewPod must default Outcome to runner.PodStatus{State: runner.StateSucceeded} when Outcome.State is empty, and default PollsBeforeDone to 1 when it is zero."
- added `r.pod-probe-returns-configured-result` (MUST): "Pod.Probe must ignore its arguments and return PodOptions.ProbeResult with a nil error."
- added `r.pod-satisfies-podrunner` (MUST): "Pod must satisfy runner.PodRunner, asserted at compile time with var _ runner.PodRunner = (*Pod)(nil)."
- added `r.pod-start-running` (MUST): "Pod.Start must allocate an id formatted as mempod-<n> from an atomic sequence and record the run in StateRunning, returning the id and a nil error."
- added `r.pod-stop-marks-failure` (MUST): "Pod.Stop must set a still-running run to runner.PodStatus{State: runner.StateFailed, Message: "stopped"}, must return nil for an unknown run id, and must return nil always."
- added `r.runs-guarded-by-mutex` (MUST): "Both Pod and Engine must guard their run maps with their own sync.Mutex, locking in Start, Observe and Stop so the fakes are safe for concurrent use."

### impl-metrics

- intent: "" -> "This context exists so that every component in the repository can expose Prometheus metrics through one small, uniform surface instead of touching the prometheus client directly. It gives each component an isolated registry plus a naming prefix, so two components in one process cannot collide on a metric name, and it makes the common collectors idempotent by name so repeated wiring is safe. It also supplies the two export paths a process needs: Render for in-process encoding and Listen for an HTTP scrape endpoint, with Server.Address and Server.Close for lifecycle control and tests that bind port zero."
- added `r.gaugefunc-exclusive` (MUST): "Registry.GaugeFunc reads its value at scrape time, is bound to one caller, and must therefore panic on any registration error instead of sharing a collector, so two gauge functions cannot take the same name."
- added `r.http-handler` (MUST): "Registry.Handler returns a promhttp handler over this registry alone, so an HTTP scrape reports only the metrics this Registry owns."
- added `r.isolated-registry` (MUST): "New creates a fresh prometheus.Registry, so a Registry never inherits metrics registered elsewhere, and Registry.Registry exposes it to callers that need the underlying collector registry."
- added `r.listen-binds-and-serves` (MUST): "Registry.Listen binds the requested TCP address, returns the bind error wrapped as "metrics: bind %s: %w", and otherwise serves Registry.Handler on that listener from a background goroutine behind the returned Server."
- added `r.prefix-names` (MUST): "Registry.Name returns the name unchanged when the prefix is empty and prefix+"_"+name otherwise, and every collector constructor must use it so all exported metric names carry the prefix."
- added `r.render-text-plain` (MUST): "Registry.Render gathers the registry and encodes every family to the writer in text/plain exposition format, wrapping a gather failure as "metrics: gather: %w" and an encode failure as "metrics: encode %s: %w" with the family name."
- added `r.server-lifecycle-nil-safe` (MUST): "Server.Address returns the bound listener address, or the empty string for a nil Server or one with no listener, and Server.Close closes the HTTP server and returns nil for a nil Server or one with no server, so both are safe on a value that was never started."
- added `r.shared-collectors-idempotent` (MUST): "Registry.Counter, Registry.Gauge and Registry.Summary must return the already registered collector when the same prefixed name is requested twice, and must panic when registration fails for any other reason, including a name registered with a different help string or a different collector type."

### impl-openbaoclient

- intent: "" -> "This context exists so that the rest of the library can treat OpenBao as one interchangeable PKI backend behind a narrow Go surface. It separates OpenBao's REST and HTTP error vocabulary from the domain types (pki.Health, pki.RootCA, pki.IntermediateCSR, pki.Role, pki.Cert, pki.CertRequest), so consumers never import the OpenBao SDK or inspect status codes, and so the same PKI workflows can be exercised against an in-process test server."
- added `r.client-construction` (MUST): "New builds a *Client from Options and returns an error when the options are not usable."
- added `r.health-mapping` (MUST): "Health queries sys/health and returns a pki.Health carrying initialized, sealed, standby and version; a transport failure is translated into a ResponseError naming GET sys/health."
- added `r.intermediate-authority` (SHOULD): "GenerateIntermediate produces a pki.IntermediateCSR for a common name, SignIntermediate signs that CSR on a root namespace and mount with a TTL and returns the chain, and SetSignedIntermediate installs the chain into the intermediate mount."
- added `r.mount-provisioning` (MUST): "EnsureMount enables a PKI engine at the given namespace and path with the given kind, and reports an error when an existing mount has a different type."
- added `r.namespace-lifecycle` (MUST): "The client can create a namespace, test whether a namespace exists, and delete a namespace, each taking a namespace path and reporting a translated error on failure."
- added `r.no-default-issuer-is-no-authority` (MUST): "An HTTP 400 whose body reports that no default issuer is currently configured means the pki mount holds no authority yet, exactly like an HTTP 404, so errors.Is(err, pki.ErrNoAuthority) is true for CASerial and for CAChain when the mount exists and has no issuer. Every other HTTP 400 stays an error that is neither pki.ErrNoAuthority nor ErrNotFound, and its ResponseError still records the method, the path, the status and the body."
- added `r.no-default-issuer-tests` (MUST): "Tests against the in-process OpenBao-shaped server cover both halves. A 400 whose body says no default issuer is currently configured makes CASerial report pki.ErrNoAuthority, and an unrelated 400 does not report pki.ErrNoAuthority and does not report ErrNotFound."
- added `r.response-error-shape` (MUST): "A failed OpenBao request is reported as a *ResponseError that records the HTTP method, the request path, the status code and the response body, and Error formats them as "openbao: %s /v1/%s -> HTTP %d: %s"."
- added `r.role-and-issuance` (MUST): "WriteRole stores a pki.Role under a named role on a mount, and Issue requests a certificate for that role from a pki.CertRequest, returning a pki.Cert."
- added `r.root-authority` (MUST): "GenerateRoot creates a root CA in the given namespace and mount from a common name and TTL and returns a pki.RootCA; CASerial and CAChain read back the authority's serial number and PEM chain."
- added `r.sentinel-error-matching` (MUST): "ResponseError.Is maps HTTP 404 to ErrNotFound, HTTP 403 to ErrForbidden and an HTTP 400 that reports no default issuer as pki.ErrNoAuthority, and returns false for every other status, so callers can use errors.Is against the sentinels."
- added `r.test-coverage` (SHOULD): "Behavior is exercised by tests against an in-process OpenBao-shaped server, including a mount type mismatch reported through ResponseError, with certificate fixtures supplied by the test certificate helper."

### impl-pkiprovisioner

- intent: "" -> "This context exists so that callers can ask for a root CA, a per-namespace intermediate authority, and leaf certificates without knowing the underlying PKI store's call sequence. It owns the ordering rules — mount the root, generate or read back the root, create the namespace, mount PKI there, sign and install an intermediate, write the role — and it owns the caching that keeps those steps from running on every request. It is the implementation behind the pki.Provisioner interface declared in abc/pki."
- added `r.authority-carries-trimmed-chain` (MUST): "The returned pki.Authority must carry the namespace path, the intermediate common name formed as path plus ".intermediate", the serial, and the CA chain with leading and trailing whitespace trimmed."
- added `r.cached-root-pem-nil-until-ensured` (MUST): "CachedRootPEM must lock the mutex and return nil when no root has been cached yet, otherwise return the cached root certificate bytes."
- added `r.delete-drops-cache-then-namespace` (MUST): "Delete must hold the mutex, drop the path from the namespace cache, and then delete the namespace through the client."
- added `r.delete-requires-path` (MUST): "Delete must reject an empty path with an error."
- added `r.ensure-authority-cached` (MUST): "EnsureAuthority must return the cached pki.Authority for the path when the expiring cache holds one at the current time, and must store a freshly provisioned authority in the cache on success."
- added `r.ensure-authority-ordering` (MUST): "Provisioning an authority must first ensure the root, then ensure the namespace, then ensure the pki mount in that namespace, and must reuse an existing intermediate serial when the namespace already has an authority, signing one only when the serial is empty."
- added `r.ensure-authority-requires-path` (MUST): "EnsureAuthority must reject an empty path with an error before taking the lock."
- added `r.ensure-root-memoized-under-mutex` (MUST): "EnsureRoot must lock the Provisioner mutex and return the already cached root when one is held, so the root is read or generated only once."
- added `r.ensure-root-mount-then-read-or-generate` (MUST): "Root provisioning must ensure the pki mount in the root namespace, read the CA serial, and treat pki.ErrNoAuthority as the absence of an authority rather than a failure; when a serial exists it must return the CA chain with that serial, otherwise it must generate a new root with the configured common name and TTL."
- added `r.intermediate-signed-by-root` (MUST): "A new intermediate must be generated in the target namespace, signed by the root namespace's mount with the configured intermediate TTL, installed back into the target namespace, and its resulting serial read back."
- added `r.issue-defaults-loopback-ips` (MUST): "Issue must substitute the loopback IP SANs 127.0.0.1 and ::1 when the caller passes no IPs."
- added `r.issue-ensures-authority` (MUST): "Issue must call EnsureAuthority for the path before requesting a certificate, so no leaf is issued against a missing intermediate."
- added `r.issue-uses-configured-role-and-ttl` (MUST): "Issue must request the certificate through the client with the configured mount, role and leaf TTL, passing the common name, alternative names and IP SANs."
- added `r.leaf-chain-joins-cert-and-ca-chain` (MUST): "LeafChain must build a string containing the leaf certificate followed by each CA chain entry on its own newline-separated line, and must return just the certificate when the chain is empty."
- added `r.new-fills-defaults` (MUST): "New must fill empty Options fields with the pki package defaults: Mount, Role, RootCommonName, RootTTL, IntermediateTTL and LeafTTL, and must set AuthorityTTL to DefaultAuthorityTTL when it is not greater than zero and Now to time.Now when nil."
- added `r.new-provides-namespace-cache` (MUST): "New must construct the Provisioner with an expiring namespace map keyed by path and bounded by the resolved AuthorityTTL."
- added `r.new-requires-a-client` (MUST): "New must return ErrNoClient when Options.Client is nil, so no Provisioner exists without a backing pki.Client."
- added `r.provisioner-satisfies-interface` (MUST): "Provisioner must satisfy pki.Provisioner, which is asserted at compile time by a blank assignment."
- added `r.role-policy` (MUST): "The leaf role written into each namespace must allow subdomains, forbid bare domains, enforce hostnames, use EC keys of 256 bits, and cap TTL at the configured leaf TTL; it must set AllowedDomains to the configured Domain only when the Domain is not empty."

### impl-policyclient

- intent: "" -> "This context exists to give the rest of the repository one place that speaks the policy engine's HTTP protocol, so callers depend on the abc/policy.Client interface and never build requests themselves. It exists because policy runs are submitted as JSON workflows to an engine endpoint and their verdicts come back as task status, and both halves need the same validation, error wrapping and output flattening."
- added `r.new-builds-a-client-with-a-bounded-timeout` (MUST): "New returns a Client whose HTTP client has a 30 second timeout, so no policy call can hang without bound."
- added `r.outputs-prefixes-verdicts-when-more-than-one-policy` (MUST): "Outputs leaves verdict keys unprefixed when exactly one policy cache key is present, and otherwise prefixes them with PolicyName of the cache key plus a slash."
- added `r.outputs-reads-policy-verdicts-from-the-cache` (MUST): "Outputs considers only cache keys with the policy/ prefix, sorted, reads each entry's result.json, skips entries that are missing or unparseable, and exports the verdict's allow flag and non-null violations."
- added `r.outputs-returns-nil-when-empty` (MUST): "Outputs returns a nil map when nothing was collected, so callers can distinguish an empty result from a populated one."
- added `r.outputs-stringifies-every-engine-output` (MUST): "Outputs maps every entry of the response's outputs object into the result map through outputs.StringifyValue."
- added `r.policy-name-reads-the-second-path-segment` (MUST): "PolicyName splits the cache key on slashes and returns the second segment when it exists and is non-empty, otherwise the literal "policy"."
- added `r.status-reads-a-task-from-an-endpoint` (MUST): "Status takes a context, an engine endpoint and a task id, and returns the policy.Task the engine reports for that task."
- added `r.submit-posts-to-request-create` (MUST): "Submit POSTs the workflow and inputs as a JSON body to the endpoint with any trailing slash trimmed, suffixed with /request/create, and with Content-Type application/json."
- added `r.submit-propagates-context-cancellation` (MUST): "When the transport call fails and the context carries an error, Submit returns the context error itself, not a wrapped transport error."
- added `r.submit-rejects-a-missing-endpoint` (MUST): "Submit returns an error naming the missing engine endpoint when the endpoint is empty, and sends no request."
- added `r.submit-requires-a-task-id` (MUST): "Submit returns the task id only when the parsed response carries a non-empty detail id, and errors when the engine returned no task id."
- added `r.submit-requires-a-valid-json-workflow` (MUST): "Submit refuses a workflow that is not valid JSON, before any request is built or sent."
- added `r.submit-treats-4xx-and-5xx-as-refusals` (MUST): "Submit returns an error carrying the status code and the trimmed response body when the engine answers with status 400 or above."

### internal-boundaries

- intent: "" -> "This context exists to keep the kcp-libs module's ABC-style layering honest by test rather than by convention. The tests read the real import graph from the Go toolchain, so any new import that crosses a layer boundary, points at internal test support from production code, or pulls a module-local package into the common leaf layer fails the build. The spec records the invariants the test file enforces and the vocabulary (layers and their ranks) it relies on, so the boundaries package can be understood without re-deriving the rules from the Go source."
- added `r.abc-imports-common-only` (MUST): "A package under the module's abc/ subtree MUST import module-local packages from the common/ subtree only, counting both Imports and TestImports; any other module-local import fails with `%s imports %s; the abc layer may import common only`."
- added `r.capture-import-graph` (MUST): "The boundaries test MUST enumerate the module's packages by running `go list -json ./...` with working directory ../.. and decoding the stream into listEntry values carrying ImportPath, Imports, TestImports and Name; a failure to run go list or decode its output MUST fail the test."
- added `r.common-is-leaf` (MUST): "A package under the module's common/ subtree MUST NOT import any other module-local package, in either Imports or TestImports; a violation fails with `%s imports %s; the common layer must be leaf packages`."
- added `r.every-package-in-a-layer` (MUST): "Every package whose import path is not .../internal/boundaries MUST resolve to one of the recognised layers common, abc, impl, factory, examples, cmd or internal; an unrecognised importer fails with `%s is not in a recognised layer`, and an unrecognised module-local import fails with `%s imports %s, which is not in a recognised layer`."
- added `r.layer-classification` (MUST): "Layer classification MUST use the first path segment after the module prefix github.com/publicdomainrelay/kcp-libs/ and MUST return empty for paths outside the module or with no such segment; rank MUST return the fixed numeric position of a known layer and -1 for an unknown one."
- added `r.one-way-layer-dependencies` (MUST): "For module-local imports, a package in any layer other than internal MUST NOT import a package whose layer rank is greater than or equal to its own; when either side is the internal layer the check is skipped. Rank order is common 0, abc 1, impl 2, factory 3, examples 4, cmd 4, internal 5, so dependencies flow common <- abc <- impl <- factory <- examples and a layer may not import its own."
- added `r.test-support-not-in-production` (MUST): "A package whose import path is outside the module's internal/ and examples/ subtrees MUST NOT import any package under the module's internal/ subtree; a violation fails with the message `%s imports %s; test support belongs in tests`."

### internal-livekcp

- intent: "" -> "This context exists so tests can run against a real kcp control plane instead of mocks, with a stable Go API for lifecycle, inspection, and object seeding. It separates tool discovery (Resolve, Missing) from cluster lifecycle (Start, Cluster.Stop) and from object manipulation (NewObject, Seed, WaitFor), so the live tier is opt-in and its prerequisites are checkable before a test starts."
- added `r.borrow-kubeconfig` (SHOULD): "When the kubeconfig environment variable is set, Start borrows that existing cluster instead of spawning a new kcp server."
- added `r.get-arbitrary-kubectl` (MUST): "Cluster.Get runs kubectl against the named logical cluster with the caller's arguments and returns the output or an error."
- added `r.kubectl-apply-manifest` (MUST): "Cluster.Kubectl applies the given manifest text to the named logical cluster and returns the command output or an error."
- added `r.logs-returns-captured-output` (MUST): "Cluster.Logs returns the captured log output of the cluster processes so startup failures can be reported with their cause."
- added `r.missing-reports-path-tools` (MUST): "Missing checks the resolved paths of "kcp" and "kubectl" with exec.LookPath and returns the names that are not found, so callers can skip or fail the live tier."
- added `r.newobject-constructor` (MUST): "NewObject builds an Object for the given namespace and name with a populated metadata block."
- added `r.object-metadata` (MUST): "Metadata and Object carry the namespace and name of a widget object so store operations can address it."
- added `r.phase-record` (MUST): "Phase records a named startup step with its Name and Duration, and Cluster.Phases returns the recorded phases in order."
- added `r.resolve-env-override` (MUST): "Resolve returns the KCP_BIN value for name "kcp" and the KUBECTL value for name "kubectl" when those environment variables are set and non-empty, and otherwise returns the name unchanged."
- added `r.seed-writes-object` (MUST): "Seed writes the given object into the resource store at the cluster, namespace, and name, and returns the store error."
- added `r.start-lifecycle` (MUST): "Start builds a Cluster rooted in a temporary directory with a free port, a generated kubeconfig, and the provider, consumer, and second-consumer workspace names, and stops the cluster if startup does not complete."
- added `r.stop-terminates-processes` (MUST): "Cluster.Stop terminates the processes the cluster spawned so a failed or finished live run leaves nothing behind."
- added `r.trace-returns-trace` (SHOULD): "Cluster.Trace returns the cluster trace as a string for test assertions and diagnostics."
- added `r.waitfor-polls-until-reached` (MUST): "WaitFor polls the store reader at the given cluster, namespace, and name until the reached predicate accepts the object, returning that object or an error when the context ends."

### internal-livekcp-cmd-livecluster

- intent: "" -> "The command exists so a developer or test harness can start a real kcp cluster as a child process and learn its coordinates from a file, instead of embedding cluster startup in Go test code. It is deliberately thin: all cluster logic lives in the livekcp package, and this file only handles flags, process signals, environment-file emission, and exit codes."
- added `r.atomic-env-file` (MUST): "The command writes `export KCP_LIBS_KUBECONFIG=<kubeconfig>` and `export KCP_LIBS_SERVER=<server>` to the env file by writing a sibling `--env` path suffixed with `.tmp` and renaming it into place, exiting with status 1 on either failure."
- added `r.env-flag-required` (MUST): "The command accepts a `--env` flag naming the file to write the environment to, and exits with status 2 and the message `livecluster: --env is required` on stderr when it is empty."
- added `r.print-trace-then-block` (MUST): "After the env file is in place, the command prints the cluster trace to stdout and blocks until the context is done."
- added `r.start-cluster-from-signal-context` (MUST): "The command starts the cluster with a context that is cancelled by SIGINT or SIGTERM, and exits with status 1 and the error prefix `livecluster:` on stderr when startup fails."
- added `r.stop-cluster-on-exit` (MUST): "The command defers Stop on the started cluster so the cluster is torn down when the process leaves main."

### internal-livekcp-livetest

- intent: "" -> "This context exists so that live tests, which need real kcp and kubectl binaries, can degrade gracefully on machines and CI runners that do not have those tools. Instead of each test repeating PATH probing and environment checks, they call livetest.Require(t) and inherit one consistent policy: skip by default, and fail loudly only when the operator has explicitly demanded the live tier by setting KCP_LIBS_REQUIRE_LIVE=1. The helper is the single switch that turns an entire test tier on and off."
- added `r.fail-when-required-and-tools-missing` (MUST): "When KCP_LIBS_REQUIRE_LIVE equals "1" and livekcp.Missing() reports at least one absent binary, Require must fail the test immediately with t.Fatalf naming the missing binaries, so a demanded live run never silently passes."
- added `r.mark-helper` (SHOULD): "Require should call t.Helper() so failures and skips are attributed to the calling test rather than to the helper."
- added `r.return-when-required-and-tools-present` (MUST): "When KCP_LIBS_REQUIRE_LIVE equals "1" and livekcp.Missing() reports nothing missing, Require must return without skipping so the calling live test runs."
- added `r.skip-naming-missing-tools` (MUST): "When the live tier is not required and livekcp.Missing() reports absent binaries, Require must skip the test with t.Skipf and a message that names those binaries."
- added `r.skip-when-not-required` (MUST): "When the live tier is not required and no binaries are missing, Require must still skip the test with t.Skipf, telling the caller to set KCP_LIBS_REQUIRE_LIVE=1 to run the live tier."

## Realization

| change | direction | phase | commit | verify | acceptance |
| --- | --- | --- | --- | --- | --- |
| factory-servicenames-s2c-c614db344601 | SpecToCode | Pending |  | 0 | - |
| factory-servicenames-s2c-d6893cea853e | SpecToCode | Succeeded | 14dcfd78 | 0 | - |
| impl-assets-s2c-f45702016a1d | SpecToCode | Failed |  | 1 | - |
| impl-assets-s2c-f45702016a1d-a2 | SpecToCode | Succeeded | f9e2ef3c | 0 | - |
| impl-assets-s2c-f81804cbe021 | SpecToCode | Succeeded | d751fc24 | 0 | - |
