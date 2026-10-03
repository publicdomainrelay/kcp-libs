package driver

import (
	"context"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

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

type Queue interface {
	Add(key Key)

	AddAfter(key Key, after time.Duration)

	AddRateLimited(key Key)

	Forget(key Key)

	Done(key Key)

	Get() (Key, bool)

	Len() int

	ShutDown()
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

func (p Policy) ConflictAfter(key Key) time.Duration {
	if p.ClampKinds[key.Kind] {
		return p.Clamp()
	}
	return p.Default()
}

const DefaultRequeueAfter = 2 * time.Second

const DefaultMinTransitionPoll = 25 * time.Millisecond

const DefaultBackstop = time.Minute
