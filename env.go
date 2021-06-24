package inspect

import (
	"os"
	"strings"
)

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/transporter/inapp"
)

const (
	EnvMetadataServiceInterface = "net.bytepowered.flux.inspect.MetadataService"
	EnvMetadataServiceMethod    = "envs"
)

func init() {
	// 注册Service
	srv := flux.ServiceSpec{
		Kind:      flux.SpecKindService,
		Protocol:  flux.ProtoInApp,
		Interface: EnvMetadataServiceInterface,
		Method:    EnvMetadataServiceMethod,
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), EnvMetadataInvokeFunc)
}

// EnvMetadataInvokeFunc 查询Env元数据信息的函数实现
func EnvMetadataInvokeFunc(ctx *flux.Context, _ flux.ServiceSpec) (interface{}, *flux.ServeError) {
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
