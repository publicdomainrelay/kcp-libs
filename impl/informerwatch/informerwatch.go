package informerwatch

import (
	"context"
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	k8scache "k8s.io/client-go/tools/cache"

	"github.com/publicdomainrelay/kcp-libs/abc/cache"
	"github.com/publicdomainrelay/kcp-libs/common/ref"
)

type Resource struct {
	Kind string

	GVR schema.GroupVersionResource
}

type Enqueue func(kind string, r ref.Ref)

type Reactor interface {
	Added(kind string, obj *unstructured.Unstructured, enqueue Enqueue)

	Updated(kind string, old, obj *unstructured.Unstructured, enqueue Enqueue)
}

type Lookup interface {
	Names(kind, index, value string) []string
}

type Options struct {
	Config *rest.Config

	Bases []string

	Resources []Resource

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
	stop := ctx.Done()
	var factories []dynamicinformer.DynamicSharedInformerFactory
	for _, base := range opts.Bases {
		client, err := factoryClient(opts.Config, base)
		if err != nil {
			return err
		}
		factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, "", nil)
		for _, resource := range opts.Resources {
			register(factory, resource, opts)
		}
		factory.Start(stop)
		factories = append(factories, factory)
	}
	for _, factory := range factories {
		factory.WaitForCacheSync(stop)
	}
	<-stop
	return nil
}

func factoryClient(config *rest.Config, base string) (dynamic.Interface, error) {
	cfg := rest.CopyConfig(config)
	cfg.Host = strings.TrimSuffix(base, "/") + ref.APIPathPrefix + "*"
	return dynamic.NewForConfig(cfg)
}

func register(factory dynamicinformer.DynamicSharedInformerFactory, resource Resource, opts Options) {
	informer := factory.ForResource(resource.GVR).Informer()
	if opts.Indexers != nil {
		_ = informer.GetIndexer().AddIndexers(toK8sIndexers(opts.Indexers))
	}
	if opts.Set != nil {
		opts.Set.Add(resource.Kind, informer.GetIndexer())
	}
	_, _ = informer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
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
