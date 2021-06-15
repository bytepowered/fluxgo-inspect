package inspect

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/toolkit"
	"github.com/bytepowered/flux/transporter/inapp"
	"sort"
)

const (
	endpointQueryApplication = "application"
	endpointQueryHttpPattern = "httpPattern"
	endpointQueryHttpMethod  = "httpMethod"

	endpointQueryVersion   = "version"
	endpointQueryServiceId = "serviceId"
	endpointQueryRpcProto  = "rpcProto"
)

const (
	EndpointMetadataServiceInterface = "net.bytepowered.flux.inspect.MetadataService"
	EndpointMetadataServiceMethod    = "QueryEndpoint"
)

type MultiEndpointFilter func(values []string, mvce *flux.MVCEndpoint) bool
type SingleEndpointFilter func(values []string, ep *flux.Endpoint) bool

// MultiEndpointFilterWrapper with value wrapper
type MultiEndpointFilterWrapper struct {
	values []string
	filter MultiEndpointFilter
}

func (w *MultiEndpointFilterWrapper) DoFilter(mvce *flux.MVCEndpoint) bool {
	return w.filter(w.values, mvce)
}

// SingleEndpointFilterWrapper with values wrapper
type SingleEndpointFilterWrapper struct {
	values []string
	filter SingleEndpointFilter
}

func (w *SingleEndpointFilterWrapper) DoFilter(ep *flux.Endpoint) bool {
	return w.filter(w.values, ep)
}

var (
	epMultiFilters  = make(map[string]MultiEndpointFilter)
	epSingleFilters = make(map[string]SingleEndpointFilter)
)

func init() {
	// 注册Service
	srv := flux.Service{
		Kind:      "flux.service/inspect/v1",
		Interface: EndpointMetadataServiceInterface,
		Method:    EndpointMetadataServiceMethod,
		Attributes: []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: flux.ProtoInApp},
		},
	}
	ext.RegisterService(srv)
	inapp.RegisterInvokeFunc(srv.ServiceID(), EndpointMetadataInvokeFunc)
	// multi filters
	epMultiFilters[endpointQueryApplication] = func(query []string, mvce *flux.MVCEndpoint) bool {
		return toolkit.MatchEqual(query, mvce.Random().Application)
	}
	epMultiFilters[endpointQueryHttpPattern] = func(query []string, mvce *flux.MVCEndpoint) bool {
		return toolkit.MatchContains(query, mvce.Random().HttpPattern)
	}
	epMultiFilters[endpointQueryHttpMethod] = func(query []string, mvce *flux.MVCEndpoint) bool {
		return toolkit.MatchEqual(query, mvce.Random().HttpMethod)
	}
	// single filter
	epSingleFilters[endpointQueryVersion] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.Version)
	}
	epSingleFilters[endpointQueryServiceId] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.ServiceId)
	}
	epSingleFilters[endpointQueryRpcProto] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.Service.RpcProto())
	}
}

// EndpointMetadataInvokeFunc 查询Endpoint元数据信息的函数实现
func EndpointMetadataInvokeFunc(ctx *flux.Context, _ flux.Service) (interface{}, *flux.ServeError) {
	// lookup
	muleps := filterMVCEndpoints(ctx)
	endpoints := filterEndpoints(ctx, muleps)
	// sort
	sort.Sort(SortableEndpoints(endpoints))
	args := extraPageArgs(ctx)
	total := len(endpoints)
	data := endpoints[limit(0, total-1, args.start):limit(0, total-1, args.end)]
	// page, pageSize
	return map[string]interface{}{
		"success":  true,
		"data":     data,
		"page":     args.page,
		"pageSize": args.pageSize,
		"total":    total,
	}, nil
}

func filterEndpoints(ctx *flux.Context, muleps []*flux.MVCEndpoint) []*flux.Endpoint {
	// Lookup filters
	filters := make([]*SingleEndpointFilterWrapper, 0, len(epSingleFilters))
	for key, filter := range epSingleFilters {
		values, ok := ctx.FormVars()[key]
		if !ok {
			continue
		}
		filters = append(filters, &SingleEndpointFilterWrapper{
			values: values,
			filter: filter,
		})
	}
	if len(filters) == 0 {
		filters = []*SingleEndpointFilterWrapper{{filter: func(_ []string, _ *flux.Endpoint) bool {
			return true
		}}}
	}
	// Data filtering
	endpoints := make([]*flux.Endpoint, 0, len(muleps))
	for _, filter := range filters {
		for _, mvce := range muleps {
			for _, ep := range mvce.Endpoints() {
				if filter.DoFilter(ep) {
					endpoints = append(endpoints, ep)
				}
			}
		}
	}
	return endpoints
}

func filterMVCEndpoints(ctx *flux.Context) []*flux.MVCEndpoint {
	// Lookup filters
	filters := make([]*MultiEndpointFilterWrapper, 0, len(epMultiFilters))
	for key, filter := range epMultiFilters {
		values, ok := ctx.FormVars()[key]
		if !ok {
			continue
		}
		filters = append(filters, &MultiEndpointFilterWrapper{
			values: values,
			filter: filter,
		})
	}
	// Data filtering
	if len(filters) == 0 {
		filters = []*MultiEndpointFilterWrapper{{
			filter: func(_ []string, _ *flux.MVCEndpoint) bool {
				return true
			}},
		}
	}
	source := ext.Endpoints()
	endpoints := make([]*flux.MVCEndpoint, 0, len(source))
	for _, filter := range filters {
		for _, ep := range source {
			if filter.DoFilter(ep) {
				endpoints = append(endpoints, ep)
			}
		}
	}
	return endpoints
}

// sort

type SortableEndpoints []*flux.Endpoint

func (s SortableEndpoints) Len() int           { return len(s) }
func (s SortableEndpoints) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortableEndpoints) Less(i, j int) bool { return keyOf(s[i]) < keyOf(s[j]) }

func keyOf(v *flux.Endpoint) string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", v.Application, v.Version, v.HttpMethod, v.HttpPattern, v.ServiceId)
}
