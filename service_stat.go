package inspect

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/transporter/inapp"
)

const (
	ServiceStatsServiceInterface = "net.bytepowered.flux.inspect.StatsService"
	ServiceStatsServiceMethod    = "services"
)

func init() {
	// 注册Service
	srv := flux.Service{
		Kind:      "flux.service/inspect/v1",
		Interface: ServiceStatsServiceInterface,
		Method:    ServiceStatsServiceMethod,
		Attributes: []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: flux.ProtoInApp},
		},
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), ServiceStatsInvokeFunc)
}

// ServiceStatsInvokeFunc 查询Service元数据统计的函数实现
func ServiceStatsInvokeFunc(_ *flux.Context, _ flux.Service) (interface{}, *flux.ServeError) {
	protos := make(map[string]int)
	total := 0
	for _, ep := range ext.Services() {
		total++
		proto := ep.RpcProto()
		if c, ok := protos[proto]; ok {
			protos[proto] = c + 1
		} else {
			protos[proto] = 1
		}
	}
	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"count":  total,
			"protos": protos,
		},
	}, nil
}
