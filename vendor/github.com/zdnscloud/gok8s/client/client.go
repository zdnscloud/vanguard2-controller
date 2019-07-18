package client

import (
	"context"
	"errors"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/zdnscloud/gok8s/client/apiutil"
	"github.com/zdnscloud/gok8s/util"
)

var (
	errMetricsServerIsNotValiable   = errors.New("metrics server isn't available")
	errDiscoveryServerIsNotValiable = errors.New("discovery server isn't available")
)

type Options struct {
	// Scheme, used to map go structs to GroupVersionKinds
	Scheme *runtime.Scheme

	// Mapper, will be used to map GroupVersionKinds to Resources
	Mapper meta.RESTMapper
}

func New(config *rest.Config, options Options) (Client, error) {
	util.Assert(config != nil, "nil rest config is provided")

	// Init a scheme if none provided
	if options.Scheme == nil {
		options.Scheme = GetDefaultScheme()
	}

	// Init a Mapper if none provided
	if options.Mapper == nil {
		mapper, err := apiutil.NewDiscoveryRESTMapper(config)
		if err != nil {
			return nil, err
		} else {
			options.Mapper = mapper
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &client{
		typedClient: typedClient{
			cache: clientCache{
				config:         config,
				scheme:         options.Scheme,
				mapper:         options.Mapper,
				codecs:         serializer.NewCodecFactory(options.Scheme),
				resourceByType: make(map[reflect.Type]*resourceMeta),
			},
			paramCodec: runtime.NewParameterCodec(options.Scheme),
		},

		unstructuredClient: unstructuredClient{
			client:     dynamicClient,
			restMapper: options.Mapper,
		},
	}, nil
}

var _ Client = &client{}

type client struct {
	typedClient        typedClient
	unstructuredClient unstructuredClient
}

func (c *client) RestClientForObject(obj runtime.Object, timeout time.Duration) (rest.Interface, error) {
	return c.typedClient.RestClientForObject(obj, timeout)
}

func (c *client) Create(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.Create(ctx, obj)
	} else {
		return c.unstructuredClient.Create(ctx, obj)
	}
}

func (c *client) Update(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.Update(ctx, obj)
	} else {
		return c.unstructuredClient.Update(ctx, obj)
	}
}

func (c *client) Patch(ctx context.Context, obj runtime.Object, typ types.PatchType, data []byte) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.Patch(ctx, obj, typ, data)
	} else {
		return c.unstructuredClient.Patch(ctx, obj, typ, data)
	}
}

func (c *client) Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOptionFunc) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.Delete(ctx, obj, opts...)
	} else {
		return c.unstructuredClient.Delete(ctx, obj, opts...)
	}
}

func (c *client) Get(ctx context.Context, key ObjectKey, obj runtime.Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.Get(ctx, key, obj)
	} else {
		return c.unstructuredClient.Get(ctx, key, obj)
	}
}

func (c *client) List(ctx context.Context, opts *ListOptions, obj runtime.Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return c.typedClient.List(ctx, opts, obj)
	} else {
		return c.unstructuredClient.List(ctx, opts, obj)
	}
}

func (c *client) Status() StatusWriter {
	return &statusWriter{client: c}
}

type statusWriter struct {
	client *client
}

var _ StatusWriter = &statusWriter{}

func (sw *statusWriter) Update(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*unstructured.Unstructured)
	if ok == false {
		return sw.client.typedClient.UpdateStatus(ctx, obj)
	} else {
		return sw.client.unstructuredClient.UpdateStatus(ctx, obj)
	}
}
