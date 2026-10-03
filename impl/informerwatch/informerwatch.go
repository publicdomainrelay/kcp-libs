package informerwatch

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	k8scache "k8s.io/client-go/tools/cache"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/common/clientlimit"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Resource struct {
	Kind string

	GVR schema.GroupVersionResource
}

type Source struct {
	Base string

	Resources []Resource
}

type Enqueue func(kind string, r ref.Ref)

type Reactor interface {
	Added(kind string, obj *unstructured.Unstructured, enqueue Enqueue)

	Updated(kind string, old, obj *unstructured.Unstructured, enqueue Enqueue)
}

type Options struct {
	Config *rest.Config

	Sources []Source

	Indexers cache.Indexers

	Set *cache.Set

	Reactor Reactor

	Enqueue Enqueue

	OnEvent func()
}

func Run(ctx context.Context, opts Options) error {
	if opts.Config == nil {
		return errors.New("informerwatch: a rest config is required")
	}
	if opts.Enqueue == nil {
		return errors.New("informerwatch: an enqueue function is required")
	}
	seen := map[string]bool{}
	for _, source := range opts.Sources {
		if source.Base == "" {
			return errors.New("informerwatch: every source needs a base URL")
		}
		if len(source.Resources) == 0 {
			return errors.New("informerwatch: every source needs at least one resource")
		}
		for _, resource := range source.Resources {
			if seen[resource.Kind] {
				return fmt.Errorf("informerwatch: %s is watched twice, and the second watcher would replace the first one's index", resource.Kind)
			}
			seen[resource.Kind] = true
		}
	}
	stop := ctx.Done()
	var factories []dynamicinformer.DynamicSharedInformerFactory
	for _, source := range opts.Sources {
		client, err := factoryClient(opts.Config, source.Base)
		if err != nil {
			return err
		}
		factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, "", nil)
		for _, resource := range source.Resources {
			if err := register(factory, resource, opts); err != nil {
				return err
			}
		}
		factory.Start(stop)
		factories = append(factories, factory)
	}
	for _, factory := range factories {
		for resource, synced := range factory.WaitForCacheSync(stop) {
			if !synced {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("informerwatch: the cache for %s did not sync", resource.Resource)
			}
		}
	}
	<-stop
	return nil
}

func factoryClient(config *rest.Config, base string) (dynamic.Interface, error) {
	cfg := clientlimit.Apply(config, 0, 0)
	cfg.Host = strings.TrimSuffix(base, "/") + ref.APIPathPrefix + "*"
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("informerwatch: build a client for %s: %w", base, err)
	}
	return client, nil
}

func register(factory dynamicinformer.DynamicSharedInformerFactory, resource Resource, opts Options) error {
	informer := factory.ForResource(resource.GVR).Informer()
	if opts.Indexers != nil {
		if err := informer.GetIndexer().AddIndexers(toK8sIndexers(opts.Indexers)); err != nil {
			return fmt.Errorf("informerwatch: index %s: %w", resource.Kind, err)
		}
	}
	if opts.Set != nil {
		opts.Set.Add(resource.Kind, informer.GetIndexer())
	}
	_, err := informer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			if opts.OnEvent != nil {
				opts.OnEvent()
			}
			enqueueOwn(resource.Kind, obj, opts.Enqueue)
			if opts.Reactor != nil {
				if u, ok := obj.(*unstructured.Unstructured); ok {
					opts.Reactor.Added(resource.Kind, u, opts.Enqueue)
				}
			}
		},
		UpdateFunc: func(oldObj, obj any) {
			if opts.OnEvent != nil {
				opts.OnEvent()
			}
			enqueueOwn(resource.Kind, obj, opts.Enqueue)
			if opts.Reactor != nil {
				old, _ := oldObj.(*unstructured.Unstructured)
				next, ok := obj.(*unstructured.Unstructured)
				if ok {
					opts.Reactor.Updated(resource.Kind, old, next, opts.Enqueue)
				}
			}
		},
	})
	return err
}

func toK8sIndexers(indexers cache.Indexers) k8scache.Indexers {
	out := make(k8scache.Indexers, len(indexers))
	for name, index := range indexers {
		out[name] = func(obj any) ([]string, error) { return index(obj) }
	}
	return out
}

func enqueueOwn(kind string, obj any, enqueue Enqueue) {
	target, ok := cache.RefOf(obj)
	if !ok {
		return
	}
	enqueue(kind, target)
}
