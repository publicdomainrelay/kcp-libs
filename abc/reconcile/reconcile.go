package reconcile

import (
	"context"
	"time"
)

type Operation string

const (
	OpStart Operation = "start"

	OpStop Operation = "stop"

	OpDelete Operation = "delete"

	OpRemoveFinalizer Operation = "remove-finalizer"
)

type Result[Status any] struct {
	Phase string

	Status Status

	Ops []Operation

	RequeueAfter time.Duration

	Delete bool

	RemoveFinalizer bool
}

func (r *Result[Status]) Add(op Operation) {
	r.Ops = append(r.Ops, op)
	switch op {
	case OpDelete:
		r.Delete = true
	case OpRemoveFinalizer:
		r.RemoveFinalizer = true
	}
}

func (r Result[Status]) Has(op Operation) bool {
	for _, candidate := range r.Ops {
		if candidate == op {
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

func Decider[Observed, Status any](f func(ctx context.Context, observed Observed) (Result[Status], error)) Reconciler[Observed, Status] {
	return Func[Observed, Status](f)
}
