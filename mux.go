package zeal

import (
	"fmt"
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/swaggerui"
	"github.com/getkin/kin-openapi/openapi3"
)

type ServeMux struct {
	*http.ServeMux
	Api *rest.API
}

func NewServeMux(apiName ...string) *ServeMux {
	name := "API"
	if len(apiName) > 0 {
		name = apiName[0]
	}
	return &ServeMux{ServeMux: http.NewServeMux(), Api: rest.NewAPI(name)}
}

type SpecOptions struct {
	Version       string
	Description   string
	StripPkgPaths []string
}

func (mux *ServeMux) CreateSpec(options SpecOptions) (*openapi3.T, error) {
	mux.Api.StripPkgPaths = options.StripPkgPaths

	spec, err := mux.Api.Spec()
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

type ServeOptions struct {
	Port           int
	Spec           *openapi3.T
	ServeSwaggerUI bool
	SwaggerPattern string
}

func (mux *ServeMux) ListenAndServe(options ServeOptions) error {
	if options.ServeSwaggerUI {
		if err := mux.ServeSwaggerUI(options.Spec, "GET "+options.SwaggerPattern); err != nil {
			return err
		}
	}

	fmt.Printf("Listening on port %v...\n", options.Port)
	if options.ServeSwaggerUI {
		fmt.Printf("Visit http://localhost:%v%v to see API definitions\n", options.Port, options.SwaggerPattern)
	}
	http.ListenAndServe(fmt.Sprintf(":%v", options.Port), mux)

	return nil
}

func (mux *ServeMux) ServeSwaggerUI(spec *openapi3.T, path string) error {
	ui, err := swaggerui.New(spec)
	if err != nil {
		return err
	}

	mux.Handle(path, ui)

	return nil
}
