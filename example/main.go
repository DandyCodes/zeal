package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DandyCodes/zeal"
	"github.com/DandyCodes/zeal/example/models"
)

var Mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

func main() {
	addRoutes(Mux)

	openApiSpecOptions := zeal.OpenAPISpecOptions{
		ZealMux:       Mux,
		Version:       "v0.1.0",
		Description:   "Example API description.",
		StripPkgPaths: []string{"main", "models", "github.com/DandyCodes/zeal"},
	}
	spec, err := zeal.CreateOpenAPISpec(openApiSpecOptions)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI spec: %v", err)
	}

	PORT := 3975
	SWAGGER_PATTERN := "/swagger-ui/"
	fmt.Printf("Visit http://localhost:%v%v to see API definitions\n", PORT, SWAGGER_PATTERN)
	zeal.ServeSwaggerUI(Mux, spec, "GET "+SWAGGER_PATTERN)

	fmt.Printf("Listening on port %v...\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%v", PORT), Mux)
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

func addRoutes(Mux *zeal.ServeMux) {
	type PostRoot struct{}
	var postRoot = zeal.Route[PostRoot](Mux)
	postRoot.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, world!")
		w.WriteHeader(http.StatusOK)
	})
	// Alternatively:
	// zeal.Route[any](Mux).HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("Hello, world!")
	// 	w.WriteHeader(http.StatusOK)
	// })

	type GetAnswer struct{ zeal.RouteResponse[int] }
	var getAnswer = zeal.Route[GetAnswer](Mux)
	getAnswer.HandleFunc("GET /answer", func(w http.ResponseWriter, r *http.Request) {
		getAnswer.Route.Response(42)
	})

	type GetMenus struct {
		zeal.RouteResponse[[]models.Menu]
	}
	var getMenus = zeal.Route[GetMenus](Mux)
	getMenus.HandleFunc("GET /menus", func(w http.ResponseWriter, r *http.Request) {
		getMenus.Route.Response(menus, http.StatusOK)
	})

	type GetMenu struct {
		zeal.RouteParams[struct {
			ID    int
			Quiet bool
		}]
		zeal.RouteResponse[models.Menu]
	}
	var getMenu = zeal.Route[GetMenu](Mux)
	getMenu.HandleFunc("GET /menus/{ID}", func(w http.ResponseWriter, r *http.Request) {
		quiet := getMenu.Route.Params().Quiet
		if !quiet {
			fmt.Println("Getting menus")
		}

		ID := getMenu.Route.Params().ID
		for i := 0; i < len(menus); i++ {
			menu := menus[i]
			if menu.ID == ID {
				getMenu.Route.Response(menu)
				return
			}
		}

		getMenu.Error(http.StatusNotFound)
	})

	type PutItem struct {
		zeal.RouteBody[models.Item]
		zeal.RouteResponse[models.Item]
	}
	var putItem = zeal.Route[PutItem](Mux)
	putItem.HandleFunc("PUT /items", func(w http.ResponseWriter, r *http.Request) {
		item := putItem.Route.Body()
		if item.Price < 0 {
			putItem.Error(http.StatusBadRequest, "Price cannot be negative")
			return
		}

		for i := range menus {
			for j := range menus[i].Items {
				if menus[i].Items[j].Name != item.Name {
					continue
				}

				menus[i].Items[j].Price = item.Price
				updatedItem := menus[i].Items[j]
				putItem.Route.Response(updatedItem)
				return
			}
		}

		menus[1].Items = append(menus[1].Items, item)
		updatedItem := menus[1].Items[len(menus[1].Items)-1]
		putItem.Route.Response(updatedItem, http.StatusCreated)
	})

	addErrRoute()
}

type PostItem struct {
	zeal.RouteParams[struct{ MenuID int }]
	zeal.RouteBody[models.Item]
}

var postItem = zeal.Route[PostItem](Mux)

func addErrRoute() {
	postItem.HandleErr("POST /items/{MenuID}", HandlePostItem)
}

func HandlePostItem(w http.ResponseWriter, r *http.Request) error {
	item := postItem.Route.Body()
	if item.Price < 0 {
		return postItem.Error(http.StatusBadRequest, "Price cannot be negative")
	}

	MenuID := postItem.Route.Params().MenuID
	for i := range menus {
		if menus[i].ID != MenuID {
			continue
		}

		menus[i].Items = append(menus[i].Items, item)
		return postItem.Status(http.StatusCreated)
	}

	return postItem.Error(http.StatusNotFound)
}
