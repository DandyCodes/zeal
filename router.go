package zeal

import (
	"log"
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/chiadapter"
	"github.com/a-h/rest/swaggerui"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	chi.Mux
	Api *rest.API
}

func NewRouter(name string) *Router {
	return &Router{Mux: *chi.NewRouter(), Api: rest.NewAPI(name)}
}

func (r *Router) CreateSpec(version string, description string) *openapi3.T {
	chiadapter.Merge(r.Api, &r.Mux)
	spec, err := r.Api.Spec()
	if err != nil {
		log.Fatalf("Failed to create spec: %v", err)
	}
	spec.Info.Version = version
	spec.Info.Description = description

	for _, queryParam := range queryParams {
		var schema *openapi3.Schema
		switch queryParam.primitiveSchemaType {
		case "boolean":
			schema = openapi3.NewBoolSchema()
		case "integer":
			schema = openapi3.NewIntegerSchema()
		case "number":
			schema = openapi3.NewFloat64Schema()
		case "string":
			schema = openapi3.NewStringSchema()
		default:
			schema = openapi3.NewStringSchema()
		}
		param := openapi3.NewQueryParameter(queryParam.name).WithSchema(schema)
		param.Required = true
		var operation *openapi3.Operation
		switch queryParam.method {
		case http.MethodGet:
			operation = spec.Paths[queryParam.pattern].Get
		case http.MethodPost:
			operation = spec.Paths[queryParam.pattern].Post
		case http.MethodPut:
			operation = spec.Paths[queryParam.pattern].Put
		case http.MethodDelete:
			operation = spec.Paths[queryParam.pattern].Delete
		}
		operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: param})
	}

	return spec
}

func (r *Router) ServeSwaggerUI(spec *openapi3.T, path string) {
	ui, err := swaggerui.New(spec)
	if err != nil {
		log.Fatalf("Failed to create swagger UI handler: %v", err)
	}
	r.Handle(path, ui)
}
