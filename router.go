package zeal

import (
	"log"

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

	for _, schemaRef := range spec.Components.Schemas {
		for propertyName := range schemaRef.Value.Properties {
			schemaRef.Value.Required = append(schemaRef.Value.Required, propertyName)
		}
	}

	for _, path := range spec.Paths {

		if path.Get != nil {
			if path.Get.Responses != nil {
				delete(path.Get.Responses, "default")
			}
		}
		if path.Post != nil {
			if path.Post.Responses != nil {
				delete(path.Post.Responses, "default")
			}
		}
		if path.Put != nil {
			if path.Put.Responses != nil {
				delete(path.Put.Responses, "default")
			}
		}
		if path.Delete != nil {
			if path.Delete.Responses != nil {
				delete(path.Delete.Responses, "default")
			}
		}
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
