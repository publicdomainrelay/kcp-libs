package reconcile

import (
	"context"
	"errors"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

var ErrGone = errors.New("reconcile: the observed object no longer exists")

type Key struct {
	Kind string

	Ref ref.Ref
}

func (k Key) String() string {
	return k.Kind + "/" + k.Ref.Key()
}

type Handler interface {
	Process(ctx context.Context, key Key) (after time.Duration, terminal bool, err error)
}

type HandlerFunc func(ctx context.Context, key Key) (time.Duration, bool, error)

func (f HandlerFunc) Process(ctx context.Context, key Key) (time.Duration, bool, error) {
	return f(ctx, key)
}

type Policy struct {
	Interval time.Duration

	MinTransitionPoll time.Duration

	ClampKinds map[string]bool
}

func (p Policy) Default() time.Duration {
	if p.Interval <= 0 {
		return DefaultRequeueAfter
	}
	return p.Interval
}

func (p Policy) Clamp() time.Duration {
	if p.MinTransitionPoll <= 0 {
		return DefaultMinTransitionPoll
	}
	return p.MinTransitionPoll
}

func (p Policy) Next(key Key, after time.Duration, terminal bool) (time.Duration, bool) {
	if terminal && after <= 0 {
		return 0, false
	}
	if after <= 0 {
		after = p.Default()
	}
	if !terminal && p.ClampKinds[key.Kind] && after > p.Clamp() {
		after = p.Clamp()
	}
	return after, true
}

func (p Policy) ConflictAfter() time.Duration {
	return p.Clamp()
}

const DefaultRequeueAfter = 2 * time.Second

const DefaultMinTransitionPoll = 25 * time.Millisecond

type Bridge[Observed, Status any] struct {
	Read func(ctx context.Context, key Key) (Observed, error)

	Decider Reconciler[Observed, Status]

	Apply func(ctx context.Context, key Key, observed Observed, result Result[Status]) error

	Terminal func(phase string) bool
}

func (b Bridge[Observed, Status]) Process(ctx context.Context, key Key) (time.Duration, bool, error) {
	observed, err := b.Read(ctx, key)
	if err != nil {
		if errors.Is(err, ErrGone) {
			return 0, true, nil
		}
		return 0, false, err
	}
	result, err := b.Decider.Reconcile(ctx, observed)
	if err != nil {
		return 0, false, err
	}
	if b.Apply != nil {
		if err := b.Apply(ctx, key, observed, result); err != nil {
			return 0, false, err
		}
	}
	return result.RequeueAfter, b.Terminal(result.Phase), nil
}
