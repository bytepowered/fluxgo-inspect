package inspect

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/transporter/inapp"
)

const (
	ServiceStatsServiceInterface = "net.bytepowered.flux.inspect.StatsService"
	ServiceStatsServiceMethod    = "services"
)

func init() {
	// 注册Service
	srv := flux.ServiceSpec{
		Kind:      flux.SpecKindService,
		Protocol:  flux.ProtoInApp,
		Interface: ServiceStatsServiceInterface,
		Method:    ServiceStatsServiceMethod,
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), ServiceStatsInvokeFunc)
}

// ServiceStatsInvokeFunc 查询Service元数据统计的函数实现
func ServiceStatsInvokeFunc(_ flux.Context, _ flux.ServiceSpec) (interface{}, *flux.ServeError) {
	protos := make(map[string]int)
	total := 0
	for _, ep := range ext.Services() {
		total++
		if c, ok := protos[ep.Protocol]; ok {
			protos[ep.Protocol] = c + 1
		} else {
			protos[ep.Protocol] = 1
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
