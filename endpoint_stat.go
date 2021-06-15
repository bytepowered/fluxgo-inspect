package inspect

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/transporter/inapp"
)

const (
	EndpointStatsServiceInterface = "net.bytepowered.flux.inspect.StatsService"
	EndpointStatsServiceMethod    = "endpoints"
)

func init() {
	// 注册Service
	srv := flux.Service{
		Kind:      "flux.service/inspect/v1",
		Interface: EndpointStatsServiceInterface,
		Method:    EndpointStatsServiceMethod,
		Attributes: []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: flux.ProtoInApp},
		},
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), EndpointStatsInvokeFunc)
}

// EndpointStatsInvokeFunc 查询Endpoint元数据统计的函数实现
func EndpointStatsInvokeFunc(_ *flux.Context, _ flux.Service) (interface{}, *flux.ServeError) {
	apps := make(map[string]int)
	count := 0
	for _, ep := range ext.Endpoints() {
		count++
		app := ep.Random().Application
		if c, ok := apps[app]; ok {
			apps[app] = c + 1
		} else {
			apps[app] = 1
		}
	}
	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"count": count,
			"apps":  apps,
		},
	}, nil
}
