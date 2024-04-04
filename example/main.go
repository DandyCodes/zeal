package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/DandyCodes/zeal"
	"github.com/DandyCodes/zeal/example/models"
)

func main() {
	mux := zeal.NewServeMux("Example API")
	addRoutes(mux)

	specOptions := zeal.SpecOptions{
		Version:       "v0.1.0",
		Description:   "Example API description.",
		StripPkgPaths: []string{"main", "models", "github.com/DandyCodes/zeal"},
	}
	spec, err := mux.CreateSpec(specOptions)
	if err != nil {
		log.Fatalf("Failed to create API: %v", err)
	}

	serveOptions := zeal.ServeOptions{
		Port:           3975,
		Spec:           spec,
		ServeSwaggerUI: true,
		SwaggerPattern: "/swagger-ui/",
	}
	mux.ListenAndServe(serveOptions)
}

var foodMenu = models.Menu{
	ID: 1,
	Items: []models.Item{
		{Name: "Steak", Price: 13.95},
		{Name: "Potatoes", Price: 3.95},
	},
}

var drinksMenu = models.Menu{
	ID: 2,
	Items: []models.Item{
		{Name: "Juice", Price: 1.25},
		{Name: "Soda", Price: 1.75},
	},
}

var menus = []models.Menu{foodMenu, drinksMenu}

func addRoutes(mux *zeal.ServeMux) {
	zeal.Handle(mux, "POST /",
		func(response zeal.Response[any], params any, body any) error {
			fmt.Println("Hello, world!")
			return response.Status(http.StatusOK)
		})

	zeal.Handle(mux, "GET /the_answer",
		func(r zeal.Response[int], p any, b any) error {
			return r.JSON(42)
		})

	zeal.Handle(mux, "GET /menus",
		func(r zeal.Response[[]models.Menu], p any, b any) error {
			return r.JSON(menus, http.StatusOK)
		})

	zeal.Handle(mux, "GET /menus/{ID}",
		func(r zeal.Response[models.Menu], p struct{ ID int }, b any) error {
			for _, menu := range menus {
				if menu.ID == p.ID {
					return r.JSON(menu)
				}
			}

			return r.Error(http.StatusNotFound)
		})

	type PutItemsParams struct {
		Quiet bool
	}
	zeal.Handle(mux, "PUT /items",
		func(r zeal.Response[models.Item], p PutItemsParams, item models.Item) error {
			if item.Price < 0 {
				return r.Error(http.StatusBadRequest, "Price cannot be negative")
			}

			for i := range menus {
				for j := range menus[i].Items {
					if menus[i].Items[j].Name != item.Name {
						continue
					}

					if !p.Quiet {
						fmt.Println("Updating item:", item)
					}
					menus[i].Items[j].Price = item.Price
					updatedItem := menus[i].Items[j]
					return r.JSON(updatedItem)
				}
			}

			if !p.Quiet {
				fmt.Println("Creating new item:", item)
			}
			menus[1].Items = append(menus[1].Items, item)
			updatedItem := menus[1].Items[len(menus[1].Items)-1]
			return r.JSON(updatedItem, http.StatusCreated)
		})

	zeal.Handle(mux, "POST /items", HandlePostItem)
}

func HandlePostItem(r zeal.Response[any], p struct{ MenuID int }, item models.Item) error {
	if item.Price < 0 {
		return r.Error(http.StatusBadRequest, "Price cannot be negative")
	}

	for i := range menus {
		if menus[i].ID != p.MenuID {
			continue
		}

		menus[i].Items = append(menus[i].Items, item)
		return r.Status(http.StatusCreated)
	}

	return r.Error(http.StatusNotFound)
}

func UseHTTPRequestAndResponse(mux *zeal.ServeMux) {
	zeal.Handle(mux, "GET /",
		func(r zeal.Response[any], p any, b any) error {
			fmt.Println(r.Request)        // *http.Request
			fmt.Println(r.ResponseWriter) // http.ResponseWriter
			return nil
		})
}

func DefineStandardRoute(mux *zeal.ServeMux) {
	mux.HandleFunc("GET /std", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware[R, P, B any](next zeal.HandlerFunc[R, P, B]) zeal.HandlerFunc[R, P, B] {
	return func(r zeal.Response[R], p P, b B) error {
		start := time.Now()

		w := &loggingResponseWriter{ResponseWriter: r.ResponseWriter, StatusCode: http.StatusOK}
		r.ResponseWriter = w

		err := next(r, p, b)

		msg := fmt.Errorf(http.StatusText(w.StatusCode))
		if err != nil {
			msg = err
		}

		log.Println(r.Request.Method, r.Request.URL.Path, w.StatusCode, msg, time.Since(start))
		return err
	}
}

func AntiDdosMiddleware[R, P, B any](next zeal.HandlerFunc[R, P, B]) zeal.HandlerFunc[R, P, B] {
	return func(r zeal.Response[R], p P, b B) error {
		if rand.Float64() < 0.33 {
			return r.Error(http.StatusTeapot, "computer says no")
		}
		return next(r, p, b)
	}
}

func MiddlewareHandle[R, P, B any](
	mux *zeal.ServeMux, urlPattern string, handlerFunc zeal.HandlerFunc[R, P, B],
) {
	loggingHandlerFunc := LoggingMiddleware(handlerFunc)
	antiDdosHandlerFunc := AntiDdosMiddleware(loggingHandlerFunc)
	zeal.Handle(mux, urlPattern, antiDdosHandlerFunc)
}

func AddMiddlewareHandleRoute(mux *zeal.ServeMux) {
	MiddlewareHandle(mux, "GET /middleware",
		func(r zeal.Response[[]models.Menu], p any, b any) error {
			if rand.Float64() < 0.33 {
				return r.Error(http.StatusInternalServerError, "an error occurred")
			} else {
				return r.JSON(menus)
			}
		})
}
