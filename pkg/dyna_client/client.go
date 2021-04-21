package dyna_client

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type DynaClient struct {
	client          dynamic.Interface
	mapper          *restmapper.DeferredDiscoveryRESTMapper
	unstructuredDEC runtime.Serializer
}

func NewDynaClient(config *rest.Config) (*DynaClient, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &DynaClient{
		client:          dyn,
		mapper:          mapper,
		unstructuredDEC: yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}, nil
}

func (d *DynaClient) UnstructuredDecode(yaml []byte) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	obj := &unstructured.Unstructured{}

	_, gvk, err := d.unstructuredDEC.Decode(yaml, nil, obj)
	if err != nil {
		return nil, nil, err
	}

	return obj, gvk, err
}

func (d *DynaClient) Apply(yaml []byte, fieldManager string) error {
	obj, gvk, err := d.UnstructuredDecode(yaml)
	if err != nil {
		return err
	}

	mapping, err := d.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = d.client.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = d.client.Resource(mapping.Resource)
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = dr.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: fieldManager,
	})

	if err != nil {
		return err
	}

	return nil
}
