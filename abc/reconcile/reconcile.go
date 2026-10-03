package reconcile

import (
	"context"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Kind string

const (
	KindStart Kind = "start"

	KindStop Kind = "stop"

	KindDelete Kind = "delete"

	KindRemoveFinalizer Kind = "remove-finalizer"

	KindCreate Kind = "create"
)

type Operation struct {
	Kind Kind

	Target ref.Ref
}

type Result[Status any] struct {
	Phase string

	Status Status

	Ops []Operation

	Clear []string

	RequeueAfter time.Duration
}

func (r *Result[Status]) Add(kind Kind) {
	r.Ops = append(r.Ops, Operation{Kind: kind})
}

func (r *Result[Status]) AddFor(kind Kind, target ref.Ref) {
	r.Ops = append(r.Ops, Operation{Kind: kind, Target: target})
}

func (r Result[Status]) Has(kind Kind) bool {
	for _, op := range r.Ops {
		if op.Kind == kind {
			return true
		}
	}
	return false
}

func (r Result[Status]) Deletes() bool {
	return r.Has(KindDelete)
}

func (r Result[Status]) ReleasesFinalizer() bool {
	return r.Has(KindRemoveFinalizer)
}

func (r Result[Status]) Cleared(field string) bool {
	for _, name := range r.Clear {
		if name == field {
			return true
		}
	}
	return false
}

type Reconciler[Observed, Status any] interface {
	Reconcile(ctx context.Context, observed Observed) (Result[Status], error)
}

type Func[Observed, Status any] func(ctx context.Context, observed Observed) (Result[Status], error)

func (f Func[Observed, Status]) Reconcile(ctx context.Context, observed Observed) (Result[Status], error) {
	return f(ctx, observed)
}
