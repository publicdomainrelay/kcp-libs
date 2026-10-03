package reconcile

import (
	"context"
	"testing"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type runStatus struct {
	Phase string

	RunID string

	Retries int32
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
		res.Clear = append(res.Clear, "runID")
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
	if !res.Cleared("runID") {
		t.Fatal("a merge patch cannot clear a field it is not told about; the run id must be in Clear")
	}
	if res.Cleared("startTime") {
		t.Fatal("Clear must name only the fields the decider cleared")
	}
}

func TestAStartCarriesNoClearedField(t *testing.T) {
	res, err := decideRun(context.Background(), runObserved{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Clear) != 0 {
		t.Fatalf("cleared = %v, want none", res.Clear)
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
