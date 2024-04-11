package zeal

import (
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/swaggerui"
	"github.com/getkin/kin-openapi/openapi3"
)

type ServeMux struct {
	*http.ServeMux
	Api *rest.API
}

func NewServeMux(mux *http.ServeMux, apiName ...string) *ServeMux {
	name := "API"
	if len(apiName) > 0 {
		name = apiName[0]
	}
	return &ServeMux{ServeMux: mux, Api: rest.NewAPI(name)}
}

type SpecOptions struct {
	ServeMux      *ServeMux
	Version       string
	Description   string
	StripPkgPaths []string
}

func NewOpenAPISpec(options SpecOptions) (*openapi3.T, error) {
	options.ServeMux.Api.StripPkgPaths = options.StripPkgPaths

	spec, err := options.ServeMux.Api.Spec()
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
		removeDefaultResponses(path.Connect)
		removeDefaultResponses(path.Delete)
		removeDefaultResponses(path.Get)
		removeDefaultResponses(path.Head)
		removeDefaultResponses(path.Options)
		removeDefaultResponses(path.Patch)
		removeDefaultResponses(path.Post)
		removeDefaultResponses(path.Put)
		removeDefaultResponses(path.Trace)
	}

	return spec, nil
}

func removeDefaultResponses(operation *openapi3.Operation) {
	if operation == nil {
		return
	}
	operation.Responses = openapi3.NewResponses(func(responses *openapi3.Responses) {
		for code, response := range operation.Responses.Map() {
			if code == "default" {
				continue
			}
			responses.Set(code, response)
		}
	})
}

func ServeSwaggerUI(mux *ServeMux, openAPISpec *openapi3.T, path string) error {
	ui, err := swaggerui.New(openAPISpec)
	if err != nil {
		return err
	}

	mux.Handle(path, ui)

	return nil
}
