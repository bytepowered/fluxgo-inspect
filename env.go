package inspect

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/transporter/inapp"
	"os"
	"strings"
)

const (
	EnvMetadataServiceInterface = "net.bytepowered.flux.inspect.MetadataService"
	EnvMetadataServiceMethod    = "QueryEnv"
)

func init() {
	// 注册Service
	srv := flux.Service{
		Kind:      "flux.service/inspect/v1",
		Interface: EnvMetadataServiceInterface,
		Method:    EnvMetadataServiceMethod,
		Attributes: []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: flux.ProtoInApp},
		},
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), EnvMetadataInvokeFunc)
}

// EnvMetadataInvokeFunc 查询Env元数据信息的函数实现
func EnvMetadataInvokeFunc(ctx *flux.Context, _ flux.Service) (interface{}, *flux.ServeError) {
	osenv := os.Environ()
	envs := make(map[string]string, len(osenv))
	for _, e := range osenv {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}
	// Resolve key
	if key := ctx.FormVar(configQueryKey); "" == key {
		return map[string]interface{}{
			"envKey": "all",
			"value":  envs,
		}, nil
	} else {
		return map[string]interface{}{
			"envKey": key,
			"value":  os.Getenv(key),
		}, nil
	}
}
