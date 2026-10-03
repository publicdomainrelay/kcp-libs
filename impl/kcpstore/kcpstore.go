package kcpstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Options struct {
	Host string

	RestConfig *rest.Config

	Transport http.RoundTripper
}

type Store struct {
	cfg *rest.Config

	http *http.Client

	codecs runtime.NegotiatedSerializer
}

func New(opts Options) (*Store, error) {
	if opts.Host == "" {
		return nil, errors.New("kcpstore: Host is required")
	}
	cfg := &rest.Config{}
	if opts.RestConfig != nil {
		cfg = rest.CopyConfig(opts.RestConfig)
	}
	cfg.Host = ref.BaseHost(opts.Host)
	cfg.ContentType = "application/json"
	cfg.AcceptContentTypes = "application/json"
	if opts.Transport != nil {
		cfg.Transport = opts.Transport
	}
	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("kcpstore: %w", err)
	}
	return &Store{
		cfg:    cfg,
		http:   httpClient,
		codecs: serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion(),
	}, nil
}

func (s *Store) Config() *rest.Config {
	return s.cfg
}

func (s *Store) HTTPClient() *http.Client {
	return s.http
}

func (s *Store) Host() string {
	return ref.BaseHost(s.cfg.Host)
}

func (s *Store) For(logicalCluster string, gv schema.GroupVersion) (rest.Interface, error) {
	cfg := rest.CopyConfig(s.cfg)
	cfg.GroupVersion = &gv
	apiPath := ref.APIPathPrefix + logicalCluster + "/apis"
	if gv.Group == "" {
		apiPath = ref.APIPathPrefix + logicalCluster + "/api"
	}
	cfg.APIPath = apiPath
	cfg.NegotiatedSerializer = s.codecs
	client, err := rest.RESTClientForConfigAndClient(cfg, s.http)
	if err != nil {
		return nil, fmt.Errorf("kcpstore: %s: %w", logicalCluster, err)
	}
	return client, nil
}

func (s *Store) ForGroupVersionResource(logicalCluster string, gv schema.GroupVersion) (rest.Interface, error) {
	return s.For(logicalCluster, gv)
}

type Resource[T any] struct {
	store *Store

	gvr schema.GroupVersionResource
}

func Of[T any](s *Store, gvr schema.GroupVersionResource) *Resource[T] {
	return &Resource[T]{store: s, gvr: gvr}
}

func (r *Resource[T]) GVR() schema.GroupVersionResource {
	return r.gvr
}

func (r *Resource[T]) client(logicalCluster string) (rest.Interface, error) {
	return r.store.For(logicalCluster, r.gvr.GroupVersion())
}

func (r *Resource[T]) GetRaw(ctx context.Context, target ref.Ref) ([]byte, error) {
	c, err := r.client(target.LogicalCluster)
	if err != nil {
		return nil, err
	}
	raw, err := c.Get().Namespace(target.Namespace).Resource(r.gvr.Resource).Name(target.Name).Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("kcpstore: read %s %s in %s: %w", r.gvr.Resource, target.Name, target.LogicalCluster, err)
	}
	return raw, nil
}

func (r *Resource[T]) Get(ctx context.Context, target ref.Ref) (*T, error) {
	raw, err := r.GetRaw(ctx, target)
	if err != nil {
		return nil, err
	}
	return decode[T](raw)
}

func (r *Resource[T]) List(ctx context.Context, logicalCluster string) ([]T, error) {
	c, err := r.client(logicalCluster)
	if err != nil {
		return nil, err
	}
	raw, err := c.Get().Resource(r.gvr.Resource).Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("kcpstore: list %s in %s: %w", r.gvr.Resource, logicalCluster, err)
	}
	return decodeList[T](raw)
}

func (r *Resource[T]) ListNamespaced(ctx context.Context, logicalCluster, namespace string) ([]T, error) {
	c, err := r.client(logicalCluster)
	if err != nil {
		return nil, err
	}
	raw, err := c.Get().Namespace(namespace).Resource(r.gvr.Resource).Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("kcpstore: list %s in %s/%s: %w", r.gvr.Resource, logicalCluster, namespace, err)
	}
	return decodeList[T](raw)
}

func (r *Resource[T]) Create(ctx context.Context, logicalCluster string, obj *T) error {
	c, err := r.client(logicalCluster)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("kcpstore: encode %s: %w", r.gvr.Resource, err)
	}
	if err := c.Post().Namespace(namespaceOf(raw)).Resource(r.gvr.Resource).Body(raw).Do(ctx).Error(); err != nil {
		return fmt.Errorf("kcpstore: create %s %s in %s: %w", r.gvr.Resource, nameOf(raw), logicalCluster, err)
	}
	return nil
}

func (r *Resource[T]) Delete(ctx context.Context, target ref.Ref) error {
	c, err := r.client(target.LogicalCluster)
	if err != nil {
		return err
	}
	err = c.Delete().Namespace(target.Namespace).Resource(r.gvr.Resource).Name(target.Name).
		Body([]byte(`{"propagationPolicy":"Background"}`)).Do(ctx).Error()
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("kcpstore: delete %s %s in %s: %w", r.gvr.Resource, target.Name, target.LogicalCluster, err)
	}
	return nil
}

func (r *Resource[T]) PatchStatus(ctx context.Context, target ref.Ref, patch []byte) error {
	c, err := r.client(target.LogicalCluster)
	if err != nil {
		return err
	}
	if err := c.Patch(mergePatch).SubResource("status").Namespace(target.Namespace).Resource(r.gvr.Resource).
		Name(target.Name).Body(patch).Do(ctx).Error(); err != nil {
		return fmt.Errorf("kcpstore: write %s status for %s in %s: %w", r.gvr.Resource, target.Name, target.LogicalCluster, err)
	}
	return nil
}

func (r *Resource[T]) Patch(ctx context.Context, target ref.Ref, patch []byte) error {
	c, err := r.client(target.LogicalCluster)
	if err != nil {
		return err
	}
	if err := c.Patch(jsonPatch).Namespace(target.Namespace).Resource(r.gvr.Resource).
		Name(target.Name).Body(patch).Do(ctx).Error(); err != nil {
		return fmt.Errorf("kcpstore: patch %s %s in %s: %w", r.gvr.Resource, target.Name, target.LogicalCluster, err)
	}
	return nil
}

func (r *Resource[T]) Finalizers(ctx context.Context, target ref.Ref) ([]string, error) {
	raw, err := r.GetRaw(ctx, target)
	if err != nil {
		return nil, err
	}
	var obj struct {
		Metadata struct {
			Finalizers []string `json:"finalizers"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("kcpstore: decode finalizers for %s: %w", target.Name, err)
	}
	return obj.Metadata.Finalizers, nil
}

func (r *Resource[T]) RemoveFinalizer(ctx context.Context, target ref.Ref, finalizer string) error {
	finalizers, err := r.Finalizers(ctx, target)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	return r.RemoveKnownFinalizer(ctx, target, finalizers, finalizer)
}

func (r *Resource[T]) RemoveKnownFinalizer(ctx context.Context, target ref.Ref, current []string, finalizer string) error {
	patch, err := removeFinalizerPatch(current, finalizer)
	if err != nil {
		return err
	}
	if patch == nil {
		return nil
	}
	return r.Patch(ctx, target, patch)
}

func (r *Resource[T]) AddFinalizer(ctx context.Context, target ref.Ref, finalizers []string) error {
	patch, err := addFinalizerPatch(finalizers)
	if err != nil {
		return err
	}
	return r.Patch(ctx, target, patch)
}

func Order[T any](items []T, name func(T) string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && name(items[j]) < name(items[j-1]); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func decode[T any](raw []byte) (*T, error) {
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("kcpstore: decode object: %w", err)
	}
	return &out, nil
}

func decodeList[T any](raw []byte) ([]T, error) {
	var list struct {
		Items []T `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("kcpstore: decode list: %w", err)
	}
	return list.Items, nil
}

func namespaceOf(raw []byte) string {
	var obj struct {
		Metadata struct {
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	_ = json.Unmarshal(raw, &obj)
	return obj.Metadata.Namespace
}

func nameOf(raw []byte) string {
	var obj struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	}
	_ = json.Unmarshal(raw, &obj)
	return obj.Metadata.Name
}

func ClusterPath(ctx context.Context, s *Store, logicalCluster string) (string, error) {
	c, err := s.For(logicalCluster, schema.GroupVersion{Group: "core.kcp.io", Version: "v1alpha1"})
	if err != nil {
		return "", err
	}
	raw, err := c.Get().Resource("logicalclusters").Name("cluster").Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("kcpstore: read logical cluster %s: %w", logicalCluster, err)
	}
	var obj struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("kcpstore: parse logical cluster %s: %w", logicalCluster, err)
	}
	path := obj.Metadata.Annotations["kcp.io/path"]
	if path == "" {
		return "", fmt.Errorf("kcpstore: logical cluster %s carries no kcp.io/path annotation", logicalCluster)
	}
	return path, nil
}

func MintServiceAccountToken(ctx context.Context, s *Store, logicalCluster, namespace, name string, ttl time.Duration) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	c, err := s.For(logicalCluster, schema.GroupVersion{Group: "", Version: "v1"})
	if err != nil {
		return "", err
	}
	seconds := int64(ttl / time.Second)
	body, err := json.Marshal(map[string]any{
		"apiVersion": "authentication.k8s.io/v1",
		"kind":       "TokenRequest",
		"spec": map[string]any{
			"audiences":         []string{},
			"expirationSeconds": seconds,
		},
	})
	if err != nil {
		return "", fmt.Errorf("kcpstore: encode the token request: %w", err)
	}
	raw, err := c.Post().Namespace(namespace).Resource("serviceaccounts").Name(name).
		SubResource("token").Body(body).Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("kcpstore: mint token for %s/%s in %s: %w", namespace, name, logicalCluster, err)
	}
	var out struct {
		Status struct {
			Token string `json:"token"`
		} `json:"status"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("kcpstore: parse the token request response: %w", err)
	}
	if out.Status.Token == "" {
		return "", fmt.Errorf("kcpstore: the TokenRequest for %s/%s in %s returned no token", namespace, name, logicalCluster)
	}
	return out.Status.Token, nil
}

func TrimCluster(host string) string {
	return strings.TrimSuffix(ref.BaseHost(host), "/")
}
