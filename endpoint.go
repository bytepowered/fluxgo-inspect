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
	EndpointMetadataServiceMethod    = "endpoints"
)

type MultiEndpointFilter func(values []string, mvce *flux.MVCEndpoint) bool
type ValueEndpointFilter func(values []string, ep *flux.Endpoint) bool

// MultiEndpointFilterWrapper with value wrapper
type MultiEndpointFilterWrapper struct {
	name   string
	values []string
	filter MultiEndpointFilter
}

func (w *MultiEndpointFilterWrapper) DoFilter(mvce *flux.MVCEndpoint) bool {
	return w.filter(w.values, mvce)
}

// ValueEndpointFilterWrapper with values wrapper
type ValueEndpointFilterWrapper struct {
	name   string
	values []string
	filter ValueEndpointFilter
}

func (w *ValueEndpointFilterWrapper) DoFilter(ep *flux.Endpoint) bool {
	return w.filter(w.values, ep)
}

var (
	epMultiFilters = make(map[string]MultiEndpointFilter)
	epValueFilters = make(map[string]ValueEndpointFilter)
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
	epValueFilters[endpointQueryApplication] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.Application)
	}
	epValueFilters[endpointQueryVersion] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.Version)
	}
	epValueFilters[endpointQueryServiceId] = func(query []string, ep *flux.Endpoint) bool {
		return toolkit.MatchEqual(query, ep.ServiceId)
	}
	epValueFilters[endpointQueryRpcProto] = func(query []string, ep *flux.Endpoint) bool {
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
	start := limit(0, total-1, args.start)
	end := limit(0, total, args.end)
	data := endpoints[start:end]
	// page, pageSize
	return map[string]interface{}{
		"success":  true,
		"data":     data,
		"page":     args.page,
		"pageSize": args.pageSize,
		"total":    total,
	}, nil
}

func filterEndpoints(ctx *flux.Context, multiends []*flux.MVCEndpoint) []*flux.Endpoint {
	// Lookup filters
	filters := make([]*ValueEndpointFilterWrapper, 0, len(epValueFilters))
	for key, filter := range epValueFilters {
		values, ok := ctx.FormVars()[key]
		if !ok || IsEmptyVars(values) {
			continue
		}
		filters = append(filters, &ValueEndpointFilterWrapper{
			name:   fmt.Sprintf("SingleKeyFilter/%s", key),
			values: values,
			filter: filter,
		})
	}
	if len(filters) == 0 {
		filters = []*ValueEndpointFilterWrapper{{name: "ValueEndpointFilter/all", filter: func(_ []string, _ *flux.Endpoint) bool {
			return true
		}}}
	}
	// Data filtering
	endpoints := make([]*flux.Endpoint, 0, len(multiends))
	isFilterMatch := func(ep *flux.Endpoint) bool {
		for _, filter := range filters {
			if !filter.DoFilter(ep) {
				return false
			}
		}
		return true
	}
	for _, multi := range multiends {
		for _, item := range multi.Endpoints() {
			if isFilterMatch(item) {
				endpoints = append(endpoints, item)
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
		if !ok || IsEmptyVars(values) {
			continue
		}
		filters = append(filters, &MultiEndpointFilterWrapper{
			name:   fmt.Sprintf("MulKeyFilter/%s", key),
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
	isFilterMatch := func(in *flux.MVCEndpoint) bool {
		for _, filter := range filters {
			if !filter.DoFilter(in) {
				return false
			}
		}
		return true
	}
	for _, src := range source {
		if isFilterMatch(src) {
			endpoints = append(endpoints, src)
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
