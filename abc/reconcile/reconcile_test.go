package reconcile

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type runStatus struct {
	Phase string `json:"phase,omitempty"`

	RunID string `json:"runID,omitempty"`

	Retries int32 `json:"retries,omitempty"`
}

type runObserved struct {
	Phase string

	Failed bool

	Running bool
}

func decideRun(_ context.Context, o runObserved) (Result[runStatus], error) {
	res := Result[runStatus]{Status: runStatus{Phase: o.Phase}}
	switch {
	case o.Failed:
		res.Phase = "Pending"
		res.Status.RunID = ""
		res.ClearFields = append(res.ClearFields, "runID")
		res.RequeueAfter = 2 * time.Second
	case o.Running:
		res.Status.RunID = "workload-1"
	case o.Phase == "":
		res.Phase = "Running"
		res.Add(KindStart)
	}
	return res, nil
}

func TestARetryCarriesTheClearedFieldSignal(t *testing.T) {
	res, err := decideRun(context.Background(), runObserved{Phase: "Running", Failed: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status.RunID != "" {
		t.Fatal("the retry must not carry the dead workload's id")
	}
	if !res.IsCleared("runID") {
		t.Fatal("a merge patch cannot clear a field it is not told about; the run id must be in Clear")
	}
	if res.IsCleared("startTime") {
		t.Fatal("Clear must name only the fields the decider cleared")
	}
}

func TestAStartCarriesNoClearedField(t *testing.T) {
	res, err := decideRun(context.Background(), runObserved{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.ClearFields) != 0 {
		t.Fatalf("cleared = %v, want none", res.ClearFields)
	}
	if !res.Has(KindStart) {
		t.Fatal("the first pass must ask for a start")
	}
}

type jobStatus struct {
	Phase string

	Runs []string
}

type jobObserved struct {
	PendingRuns int

	RunningRuns []string
}

func decideJob(_ context.Context, o jobObserved) (Result[jobStatus], error) {
	res := Result[jobStatus]{Phase: "Running", Status: jobStatus{Phase: "Running"}}
	if o.PendingRuns > 0 {
		for i := range o.PendingRuns {
			res.AddFor(KindCreate, ref.New("root:alice", "default", "job-1-run-"+string(rune('a'+i))))
		}
	}
	for _, name := range o.RunningRuns {
		res.AddFor(KindStop, ref.New("root:alice", "default", name))
	}
	return res, nil
}

func TestAJobCarriesActionPayloadsForItsChildren(t *testing.T) {
	res, err := decideJob(context.Background(), jobObserved{PendingRuns: 2, RunningRuns: []string{"old-1"}})
	if err != nil {
		t.Fatal(err)
	}
	var created, stopped []ref.Ref
	for _, op := range res.Ops {
		switch op.Kind {
		case KindCreate:
			created = append(created, op.Target)
		case KindStop:
			stopped = append(stopped, op.Target)
		}
	}
	if len(created) != 2 || created[0].Name != "job-1-run-a" || created[1].Name != "job-1-run-b" {
		t.Fatalf("created = %v", created)
	}
	if len(stopped) != 1 || stopped[0].Name != "old-1" {
		t.Fatalf("stopped = %v", stopped)
	}
}

func TestLifecycleOperationsAreDerivedNotRestated(t *testing.T) {
	res := Result[jobStatus]{}
	if res.Deletes() || res.ReleasesFinalizer() {
		t.Fatal("an empty result asks for nothing")
	}
	res.Add(KindDelete)
	if !res.Deletes() {
		t.Fatal("Delete must follow the operation list, not a second field that can disagree with it")
	}
	res.Add(KindRemoveFinalizer)
	if !res.ReleasesFinalizer() {
		t.Fatal("RemoveFinalizer must follow the operation list")
	}
}

func TestOperationOnTheObservedObjectHasNoTarget(t *testing.T) {
	res := Result[runStatus]{}
	res.Add(KindStop)
	if res.Ops[0].Target != (ref.Ref{}) {
		t.Fatalf("an operation on the observed object carries no target, got %+v", res.Ops[0].Target)
	}
}

func TestFuncAdaptsAFunction(t *testing.T) {
	reconciler := Func[runObserved, runStatus](decideRun)
	res, err := reconciler.Reconcile(context.Background(), runObserved{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Has(KindStart) {
		t.Fatal("Func must call the function it wraps")
	}
}

func TestBridgeReadsDecidesAndApplies(t *testing.T) {
	applied := false
	bridge := Bridge[runObserved, runStatus]{
		Read: func(_ context.Context, key Key) (runObserved, error) {
			if key.Ref.Name == "gone" {
				return runObserved{}, ErrGone
			}
			return runObserved{Running: true}, nil
		},
		Decider: Func[runObserved, runStatus](decideRun),
		Apply: func(_ context.Context, _ Key, observed runObserved, result Result[runStatus]) error {
			if !observed.Running {
				t.Fatal("Apply must receive what Read read")
			}
			applied = true
			return nil
		},
		Terminal: func(phase string) bool { return phase == "Succeeded" || phase == "Failed" },
	}

	key := Key{Kind: "run", Ref: ref.New("root:alice", "default", "one")}
	after, terminal, err := bridge.Process(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if terminal || after != 0 {
		t.Fatalf("a running run is not terminal: (%v, %v)", after, terminal)
	}

	gone := Key{Kind: "run", Ref: ref.New("root:alice", "default", "gone")}
	_, terminal, err = bridge.Process(context.Background(), gone)
	if err != nil {
		t.Fatal(err)
	}
	if !terminal {
		t.Fatal("an object that is gone must not be retried")
	}
	if !applied {
		t.Fatal("the bridge must apply the result")
	}
}

func TestThePatchCarriesTheClearedFieldAsNull(t *testing.T) {
	res, err := decideRun(context.Background(), runObserved{Phase: "Running", Failed: true})
	if err != nil {
		t.Fatal(err)
	}
	patch, err := Patch(res)
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Status map[string]any `json:"status"`
	}
	if err := json.Unmarshal(patch, &body); err != nil {
		t.Fatal(err)
	}
	value, present := body.Status["runID"]
	if !present {
		t.Fatalf("a merge patch cannot clear a field it omits: %s", patch)
	}
	if value != nil {
		t.Fatalf("the cleared field must be null, got %v", value)
	}
	if body.Status["phase"] != "Running" {
		t.Fatalf("the rest of the status must survive: %s", patch)
	}
}

func TestThePatchOmitsWhatTheStatusDidNotCarry(t *testing.T) {
	patch, err := Patch(Result[runStatus]{Status: runStatus{Phase: "Running"}})
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Status map[string]any `json:"status"`
	}
	if err := json.Unmarshal(patch, &body); err != nil {
		t.Fatal(err)
	}
	if _, present := body.Status["runID"]; present {
		t.Fatalf("an untouched field must not be in the patch: %s", patch)
	}
}

func TestBridgeReportsAFailedApply(t *testing.T) {
	bridge := Bridge[runObserved, runStatus]{
		Read:     func(context.Context, Key) (runObserved, error) { return runObserved{}, nil },
		Decider:  Func[runObserved, runStatus](decideRun),
		Apply:    func(context.Context, Key, runObserved, Result[runStatus]) error { return errApply },
		Terminal: func(string) bool { return false },
	}
	if _, _, err := bridge.Process(context.Background(), Key{}); !errors.Is(err, errApply) {
		t.Fatalf("err = %v, want the apply failure to surface", err)
	}
}

var errApply = errors.New("the status write failed")
