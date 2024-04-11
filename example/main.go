package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DandyCodes/zeal"
	"github.com/DandyCodes/zeal/example/models"
)

var mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

func main() {
	addRoutes(mux)

	specOptions := zeal.SpecOptions{
		ServeMux:      mux,
		Version:       "v0.1.0",
		Description:   "Example API description.",
		StripPkgPaths: []string{"main", "models", "github.com/DandyCodes/zeal"},
	}
	openAPISpec, err := zeal.NewOpenAPISpec(specOptions)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI spec: %v", err)
	}

	port := 3975
	swaggerPattern := "/swagger-ui/"
	fmt.Printf("Visit http://localhost:%v%v to see API definitions\n", port, swaggerPattern)
	zeal.ServeSwaggerUI(mux, openAPISpec, "GET "+swaggerPattern)

	fmt.Printf("Listening on port %v...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
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
	var route = zeal.NewRoute[zeal.Route](mux)
	route.HandleFunc("POST /hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, world!")
	})

	type GetAnswer struct {
		zeal.Route
		zeal.HasResponse[int]
	}
	var getAnswer = zeal.NewRoute[GetAnswer](mux)
	getAnswer.HandleFunc("GET /answer", func(w http.ResponseWriter, r *http.Request) {
		getAnswer.Response(42)
	})

	type GetMenus struct {
		zeal.Route
		zeal.HasResponse[[]models.Menu]
	}
	var getMenus = zeal.NewRoute[GetMenus](mux)
	getMenus.HandleFunc("GET /menus", func(w http.ResponseWriter, r *http.Request) {
		getMenus.Response(menus)
	})

	type DeleteMenu struct {
		zeal.Route
		zeal.HasParams[struct {
			ID    int
			Quiet bool
		}]
	}
	var deleteMenu = zeal.NewRoute[DeleteMenu](mux)
	deleteMenu.HandleFunc("DELETE /menus/{ID}", func(w http.ResponseWriter, r *http.Request) {
		if !deleteMenu.Params().Quiet {
			fmt.Println("Deleting menu")
		}

		for i := 0; i < len(menus); i++ {
			if menus[i].ID == deleteMenu.Params().ID {
				menus = append(menus[:i], menus[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	type PutItem struct {
		zeal.Route
		zeal.HasBody[models.Item]
	}
	var putItem = zeal.NewRoute[PutItem](mux)
	putItem.HandleFunc("PUT /items", func(w http.ResponseWriter, r *http.Request) {
		item := putItem.Body()
		if item.Price < 0 {
			http.Error(w, "Price cannot be negative", http.StatusBadRequest)
			return
		}

		for i := range menus {
			for j := range menus[i].Items {
				if menus[i].Items[j].Name == item.Name {
					menus[i].Items[j].Price = item.Price
					return
				}
			}
		}

		menus[0].Items = append(menus[0].Items, item)
		w.WriteHeader(http.StatusCreated)
	})

	addOuterScopeRoute()
}

type PostItem struct {
	zeal.Route
	zeal.HasParams[struct{ MenuID int }]
	zeal.HasBody[models.Item]
	zeal.HasResponse[models.Item]
}

var postItem = zeal.NewRoute[PostItem](mux)

func addOuterScopeRoute() {
	postItem.HandleFuncErr("POST /items/{MenuID}", HandlePostItem)
}

func HandlePostItem(w http.ResponseWriter, r *http.Request) error {
	item := postItem.Body()
	if item.Price < 0 {
		return zeal.Error(w, "Price cannot be negative", http.StatusBadRequest)
	}

	for i := range menus {
		if menus[i].ID == postItem.Params().MenuID {
			menus[i].Items = append(menus[i].Items, item)
			return postItem.Response(item, http.StatusCreated)
		}
	}

	return zeal.WriteHeader(w, http.StatusNotFound)
}
