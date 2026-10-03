package store

import (
	"context"
	"time"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Reader[T any] interface {
	Get(ctx context.Context, r ref.Ref) (*T, error)

	List(ctx context.Context, logicalCluster string) ([]T, error)
}

type Writer[T any] interface {
	Create(ctx context.Context, logicalCluster string, obj *T) error

	Delete(ctx context.Context, r ref.Ref) error

	PatchStatus(ctx context.Context, r ref.Ref, patch []byte) error

	RemoveFinalizer(ctx context.Context, r ref.Ref, finalizer string) error
}

type Resource[T any] interface {
	Reader[T]

	Writer[T]
}

type TokenMinter interface {
	MintServiceAccountToken(ctx context.Context, logicalCluster, namespace, name string, ttl time.Duration) (string, error)
}

type PathResolver interface {
	ClusterPath(ctx context.Context, logicalCluster string) (string, error)
}
