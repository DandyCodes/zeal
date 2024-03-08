package zeal

import (
	"log"
	"net/http"

	"github.com/a-h/rest"
	"github.com/a-h/rest/swaggerui"
	"github.com/getkin/kin-openapi/openapi3"
)

type Router struct {
	http.ServeMux
	Api *rest.API
}

func NewRouter(name string) *Router {
	return &Router{ServeMux: *http.NewServeMux(), Api: rest.NewAPI(name)}
}

func (router *Router) CreateSpec(version string, description string) *openapi3.T {
	spec, err := router.Api.Spec()
	if err != nil {
		log.Fatalf("Failed to create spec: %v", err)
	}
	spec.Info.Version = version
	spec.Info.Description = description

	for _, schemaRef := range spec.Components.Schemas {
		for propertyName := range schemaRef.Value.Properties {
			schemaRef.Value.Required = append(schemaRef.Value.Required, propertyName)
		}
	}

	for _, path := range spec.Paths.Map() {
		removeDefaultResponses(path.Get)
		removeDefaultResponses(path.Post)
		removeDefaultResponses(path.Put)
		removeDefaultResponses(path.Delete)
	}

	return spec
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

func (router *Router) ServeSwaggerUI(spec *openapi3.T, path string) {
	ui, err := swaggerui.New(spec)
	if err != nil {
		log.Fatalf("Failed to create swagger UI handler: %v", err)
	}
	router.Handle(path, ui)
}
