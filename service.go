package inspect

import (
	"fmt"
	"sort"
)

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/toolkit"
	"github.com/bytepowered/fluxgo/pkg/transporter/inapp"
)

const (
	serviceQueryApplication = "application"
	serviceQueryInterface   = "interface"
	serviceQueryMethod      = "method"
	serviceQueryRpcProto    = "rpcProto"
)

const (
	ServiceMetadataServiceInterface = "net.bytepowered.flux.inspect.MetadataService"
	ServiceMetadataServiceMethod    = "services"
)

type ServiceFilter func(values []string, ep *flux.ServiceSpec) bool

// ServiceFilterWrapper with values wrapper
type ServiceFilterWrapper struct {
	values []string
	filter ServiceFilter
}

func (w *ServiceFilterWrapper) DoFilter(srv *flux.ServiceSpec) bool {
	return w.filter(w.values, srv)
}

var (
	serviceFilters = make(map[string]ServiceFilter)
)

func init() {
	// 注册Service
	srv := flux.ServiceSpec{
		Kind:      flux.SpecKindService,
		Protocol:  flux.ProtoInApp,
		Interface: ServiceMetadataServiceInterface,
		Method:    ServiceMetadataServiceMethod,
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), ServiceMetadataInvokeFunc)
	// single filter
	serviceFilters[serviceQueryApplication] = func(query []string, ep *flux.ServiceSpec) bool {
		return true // TODO 需要底层元数据模型支持
	}
	serviceFilters[serviceQueryInterface] = func(query []string, ep *flux.ServiceSpec) bool {
		return toolkit.MatchPrefix(query, ep.Interface)
	}
	serviceFilters[serviceQueryMethod] = func(query []string, ep *flux.ServiceSpec) bool {
		return toolkit.MatchPrefix(query, ep.Method)
	}
	serviceFilters[serviceQueryRpcProto] = func(query []string, ep *flux.ServiceSpec) bool {
		return toolkit.MatchEqual(query, ep.Protocol)
	}
}

// ServiceMetadataInvokeFunc 查询Service元数据信息的函数实现
func ServiceMetadataInvokeFunc(ctx *flux.Context, _ flux.ServiceSpec) (interface{}, *flux.ServeError) {
	// lookup
	services := filterServices(ctx)
	// sort
	sort.Sort(SortableServices(services))
	total := len(services)
	args := extraPageArgs(ctx)
	data := services[limit(0, total-1, args.start):limit(0, total-1, args.end)]
	// page, pageSize
	return map[string]interface{}{
		"success":  true,
		"data":     data,
		"page":     args.page,
		"pageSize": args.pageSize,
		"total":    total,
	}, nil
}

func filterServices(ctx *flux.Context) []flux.ServiceSpec {
	// Lookup filters
	filters := make([]*ServiceFilterWrapper, 0, len(serviceFilters))
	for key, filter := range serviceFilters {
		values, ok := ctx.FormVars()[key]
		if !ok || IsEmptyVars(values) {
			continue
		}
		filters = append(filters, &ServiceFilterWrapper{
			values: values,
			filter: filter,
		})
	}
	if len(filters) == 0 {
		filters = []*ServiceFilterWrapper{{filter: func(_ []string, _ *flux.ServiceSpec) bool {
			return true
		}}}
	}
	// Data filtering
	source := ext.Services()
	services := make([]flux.ServiceSpec, 0, len(source))
	isFilterMatch := func(srv *flux.ServiceSpec) bool {
		for _, filter := range filters {
			if !filter.DoFilter(srv) {
				return false
			}
		}
		return true
	}
	for _, srv := range source {
		if isFilterMatch(&srv) {
			services = append(services, srv)
		}
	}
	return services
}

// sort

type SortableServices []flux.ServiceSpec

func (s SortableServices) Len() int           { return len(s) }
func (s SortableServices) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortableServices) Less(i, j int) bool { return serviceKeyOf(s[i]) < serviceKeyOf(s[j]) }

func serviceKeyOf(v flux.ServiceSpec) string {
	return fmt.Sprintf("%s,%s", v.Interface, v.Method)
}
