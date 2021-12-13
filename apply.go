package main

import (
	"context"
	"io"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func Apply(ctx context.Context, cfg *rest.Config, objects []*unstructured.Unstructured) error {
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}

	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	if err != nil {
		return err
	}
	for _, o := range objects {
		err = createOrUpdateResource(ctx, client, rm, o)
		if err != nil {
			return err
		}
	}
	return nil
}

func ApplyResource(ctx context.Context, cfg *rest.Config, object *unstructured.Unstructured) error {
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}

	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	if err != nil {
		return err
	}

	return createOrUpdateResource(ctx, client, rm, object)
}

func createOrUpdateResource(ctx context.Context, client dynamic.Interface, rm *restmapper.DeferredDiscoveryRESTMapper, u *unstructured.Unstructured) error {
	gvk := u.GroupVersionKind()
	mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}
	rc := client.Resource(mapping.Resource)

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		_, err = rc.Namespace(u.GetNamespace()).Get(ctx, u.GetName(), metav1.GetOptions{})
	} else {
		_, err = rc.Get(ctx, u.GetName(), metav1.GetOptions{})
	}
	if errors.IsNotFound(err) {
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			_, err = rc.Namespace(u.GetNamespace()).Create(ctx, u, metav1.CreateOptions{})
		} else {
			_, err = rc.Create(ctx, u, metav1.CreateOptions{})
		}
	} else if err == nil {
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			_, err = rc.Namespace(u.GetNamespace()).Update(ctx, u, metav1.UpdateOptions{})
		} else {
			_, err = rc.Update(ctx, u, metav1.UpdateOptions{})
		}
	}

	return err
}

func Template(out io.Writer, resources []*unstructured.Unstructured) error {
	o := yaml.NewEncoder(out)
	o.SetIndent(2)
	for _, r := range resources {
		err := o.Encode(&r.Object)
		if err != nil {
			return err
		}
	}
	return nil
}
