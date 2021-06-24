package inspect

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/transporter/inapp"
)

const (
	configQueryNamespace = "namespace"
	configQueryKey       = "key"
)

const (
	ConfigMetadataServiceInterface = "net.bytepowered.flux.inspect.MetadataService"
	ConfigMetadataServiceMethod    = "configs"
)

func init() {
	// 注册Service
	srv := flux.ServiceSpec{
		Kind:      flux.SpecKindService,
		Protocol:  flux.ProtoInApp,
		Interface: ConfigMetadataServiceInterface,
		Method:    ConfigMetadataServiceMethod,
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), ConfigMetadataInvokeFunc)
}

// ConfigMetadataInvokeFunc 查询Config元数据信息的函数实现
func ConfigMetadataInvokeFunc(ctx *flux.Context, _ flux.ServiceSpec) (interface{}, *flux.ServeError) {
	root := flux.NewRootConfiguration()
	// Namespaces
	ns := ctx.FormVar(configQueryNamespace)
	switch ns {
	case "all", "":
		// root = root
	default:
		root = flux.NewConfiguration(ns)
	}
	// Resolve key
	if key := ctx.FormVar(configQueryKey); "" == key {
		return map[string]interface{}{
			"namespace": "all",
			"key":       key,
			"value":     root.ToStringMap(),
		}, nil
	} else {
		return map[string]interface{}{
			"namespace": ns,
			"key":       key,
			"value":     root.Get(key),
		}, nil
	}
}
