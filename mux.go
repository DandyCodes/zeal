package zeal

import (
	"net/http"
	"strings"

	"github.com/a-h/rest"
	"github.com/a-h/rest/swaggerui"
	"github.com/getkin/kin-openapi/openapi3"
)

type ZealMux struct {
	*http.ServeMux
	Api *rest.API
}

func NewZealMux(mux *http.ServeMux, apiName ...string) *ZealMux {
	name := "API"
	if len(apiName) > 0 {
		name = apiName[0]
	}
	return &ZealMux{ServeMux: mux, Api: rest.NewAPI(name)}
}

type SpecOptions struct {
	ZealMux       *ZealMux
	Version       string
	Description   string
	StripPkgPaths []string
}

func NewOpenAPISpec(options SpecOptions) (*openapi3.T, error) {
	options.ZealMux.Api.StripPkgPaths = options.StripPkgPaths

	spec, err := options.ZealMux.Api.Spec()
	if err != nil {
		return nil, err
	}

	spec.Info.Version = options.Version
	spec.Info.Description = options.Description

	for _, schemaRef := range spec.Components.Schemas {
		for propertyName := range schemaRef.Value.Properties {
			schemaRef.Value.Required = append(schemaRef.Value.Required, propertyName)
		}
	}

	for _, path := range spec.Paths.Map() {
		prepareForConsumption(path.Connect)
		prepareForConsumption(path.Delete)
		prepareForConsumption(path.Get)
		prepareForConsumption(path.Head)
		prepareForConsumption(path.Options)
		prepareForConsumption(path.Patch)
		prepareForConsumption(path.Post)
		prepareForConsumption(path.Put)
		prepareForConsumption(path.Trace)
	}

	return spec, nil
}

func prepareForConsumption(operation *openapi3.Operation) {
	if operation == nil {
		return
	}
	removeDefaultResponses(operation)
	requireRequestBody(operation)
}

func removeDefaultResponses(operation *openapi3.Operation) {
	operation.Responses = openapi3.NewResponses(func(newResponses *openapi3.Responses) {
		for code, response := range operation.Responses.Map() {
			if code == "default" {
				continue
			}
			newResponses.Set(code, response)
		}
	})
}

func requireRequestBody(operation *openapi3.Operation) {
	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return
	}
	operation.RequestBody.Value.Required = true
}

func ServeSwaggerUI(mux *ZealMux, openAPISpec *openapi3.T, path string) error {
	ui, err := swaggerui.New(openAPISpec)
	if err != nil {
		return err
	}

	mux.Handle(path, ui)

	return nil
}

func StripPrefix(prefix string, h *ZealMux) *ZealMux {
	h.Handle("/", http.StripPrefix(prefix, h))
	return h
}

func (m *ZealMux) Handle(pattern string, handler http.Handler) {
	switch sHandler := handler.(type) {
	case *ZealMux:
		for _, methodToRoute := range sHandler.Api.Routes {
			for _, route := range methodToRoute {
				mergeRoute(strings.TrimSuffix(pattern, "/"), m.Api, route)
			}
		}
		m.ServeMux.Handle(pattern, sHandler)
	default:
		m.ServeMux.Handle(pattern, sHandler)
	}
}

func mergeRoute(prefix string, api *rest.API, r *rest.Route) {
	toUpdate := api.Route(string(r.Method), prefix+string(r.Pattern))
	mergeMap(toUpdate.Params.Path, r.Params.Path)
	mergeMap(toUpdate.Params.Query, r.Params.Query)
	if toUpdate.Models.Request.Type == nil {
		toUpdate.Models.Request = r.Models.Request
	}
	mergeMap(toUpdate.Models.Responses, r.Models.Responses)
}

func mergeMap[TKey comparable, TValue any](into, from map[TKey]TValue) {
	for kf, vf := range from {
		_, ok := into[kf]
		if !ok {
			into[kf] = vf
		}
	}
}
