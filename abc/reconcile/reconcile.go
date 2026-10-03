package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
	"github.com/publicdomainrelay/kcp-libs/common/statuspatch"
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

	ClearFields []string

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

func (r Result[Status]) IsCleared(field string) bool {
	return slices.Contains(r.ClearFields, field)
}

func Patch[Status any](result Result[Status]) ([]byte, error) {
	fields := map[string]any{}
	body, err := json.Marshal(result.Status)
	if err != nil {
		return nil, fmt.Errorf("reconcile: encode the status: %w", err)
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("reconcile: decode the status: %w", err)
	}
	for _, name := range result.ClearFields {
		fields[name] = nil
	}
	patch, err := statuspatch.Merge(fields)
	if err != nil {
		return nil, fmt.Errorf("reconcile: %w", err)
	}
	return patch, nil
}

type Reconciler[Observed, Status any] interface {
	Reconcile(ctx context.Context, observed Observed) (Result[Status], error)
}

type Func[Observed, Status any] func(ctx context.Context, observed Observed) (Result[Status], error)

func (f Func[Observed, Status]) Reconcile(ctx context.Context, observed Observed) (Result[Status], error) {
	return f(ctx, observed)
}
