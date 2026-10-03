package exportwatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/publicdomainrelay/kcp-libs/common/clientlimit"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

var EndpointSliceGVR = schema.GroupVersionResource{Group: "apis.kcp.io", Version: "v1alpha1", Resource: "apiexportendpointslices"}

var ErrReady = errors.New("exportwatch: the APIExport virtual workspace endpoints are ready")

const DefaultPoll = 2 * time.Second

type Endpoints map[string][]string

type Options struct {
	Config *rest.Config

	Host string

	ProviderWorkspace string

	Exports []string

	Log *slog.Logger

	Poll time.Duration
}

func (o Options) host() string {
	if o.Host != "" {
		return ref.BaseHost(o.Host)
	}
	return ref.BaseHost(o.Config.Host)
}

func Client(opts Options) (dynamic.Interface, error) {
	if opts.Config == nil {
		return nil, errors.New("exportwatch: a rest config is required")
	}
	cfg := clientlimit.Apply(opts.Config, 0, 0)
	cfg.Host = opts.host() + ref.APIPathPrefix + opts.ProviderWorkspace
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("exportwatch: build a client for %s: %w", opts.ProviderWorkspace, err)
	}
	return client, nil
}

func Discover(ctx context.Context, opts Options) (Endpoints, error) {
	client, err := Client(opts)
	if err != nil {
		return nil, err
	}
	list, err := client.Resource(EndpointSliceGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("exportwatch: list the endpoint slices: %w", err)
	}
	return FromList(list), nil
}

func FromList(list *unstructured.UnstructuredList) Endpoints {
	out := Endpoints{}
	for i := range list.Items {
		item := &list.Items[i]
		export, _, _ := unstructured.NestedString(item.Object, "spec", "export", "name")
		if export == "" {
			continue
		}
		entries, _, _ := unstructured.NestedSlice(item.Object, "status", "endpoints")
		for _, raw := range entries {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			url, _ := entry["url"].(string)
			if url != "" {
				out[export] = append(out[export], url)
			}
		}
	}
	return out
}

func (e Endpoints) Ready(exports ...string) bool {
	for _, export := range exports {
		if len(e[export]) == 0 {
			return false
		}
	}
	return true
}

func Await(ctx context.Context, opts Options) (Endpoints, error) {
	if len(opts.Exports) == 0 {
		return nil, errors.New("exportwatch: at least one export name is required")
	}
	ready := func(endpoints Endpoints) bool { return endpoints.Ready(opts.Exports...) }
	endpoints, err := Discover(ctx, opts)
	if err == nil && ready(endpoints) {
		return endpoints, nil
	}
	logEndpointWait(opts, endpoints, err)
	poll := opts.Poll
	if poll <= 0 {
		poll = DefaultPoll
	}
	for {
		werr := WatchEndpointSlices(ctx, opts, poll)
		if werr != nil && !errors.Is(werr, ErrReady) && opts.Log != nil {
			opts.Log.Warn("exportwatch: waiting for the APIExport virtual workspace endpoints", "err", werr)
		}
		endpoints, err = Discover(ctx, opts)
		if err == nil && ready(endpoints) {
			return endpoints, nil
		}
		if err != nil && opts.Log != nil {
			opts.Log.Warn("exportwatch: listing the APIExport virtual workspace endpoints", "err", err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(poll):
		}
	}
}

func WatchEndpointSlices(ctx context.Context, opts Options, wait time.Duration) error {
	bounded, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	client, err := Client(opts)
	if err != nil {
		return err
	}
	w, err := client.Resource(EndpointSliceGVR).Watch(bounded, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer w.Stop()
	for {
		select {
		case <-bounded.Done():
			return nil
		case event, ok := <-w.ResultChan():
			if !ok {
				return nil
			}
			switch event.Type {
			case watch.Added, watch.Modified, watch.Deleted:
				endpoints, err := Discover(bounded, opts)
				if err == nil && endpoints.Ready(opts.Exports...) {
					return ErrReady
				}
			}
		}
	}
}

func logEndpointWait(opts Options, endpoints Endpoints, err error) {
	if opts.Log == nil {
		return
	}
	if err != nil {
		opts.Log.Warn("exportwatch: waiting for the APIExport virtual workspace endpoints", "err", err)
		return
	}
	fields := make([]any, 0, len(opts.Exports)*2)
	for _, export := range opts.Exports {
		fields = append(fields, export, len(endpoints[export]))
	}
	opts.Log.Info("exportwatch: waiting for the APIExport virtual workspace endpoints", fields...)
}

func Paths(endpoints Endpoints, export string) []string {
	return endpoints[export]
}

func Counts(endpoints Endpoints) map[string]int {
	out := make(map[string]int, len(endpoints))
	for export, urls := range endpoints {
		out[export] = len(urls)
	}
	return out
}
