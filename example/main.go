package main

import (
	"fmt"
	"net/http"

	"github.com/DandyCodes/zeal"
	"github.com/DandyCodes/zeal/example/models"
)

func main() {
	router := zeal.NewRouter("API")

	addRoutes(router)

	router.Api.StripPkgPaths = []string{"main", "models", "github.com/DandyCodes/zeal"}
	spec := router.CreateSpec("v1.0.0", "Spec")
	router.ServeSwaggerUI(spec, "GET /swagger-ui/")

	fmt.Println("Listening on port 3000...")
	fmt.Println("Visit http://localhost:3000/swagger-ui to see API definitions")
	http.ListenAndServe(":3000", router)
}

var foodMenu = models.Menu{
	ID:    1,
	Items: []models.Item{{Name: "Steak", Price: 13.95}, {Name: "Potatoes", Price: 3.95}},
}

var drinksMenu = models.Menu{
	ID:    2,
	Items: []models.Item{{Name: "Juice", Price: 1.25}, {Name: "Soda", Price: 1.75}},
}

var menus = []models.Menu{foodMenu, drinksMenu}

func addRoutes(router *zeal.Router) {
	zeal.Route(router, "GET /",
		func(w zeal.ResponseWriter[any], r *zeal.Request[any]) {
			w.Write([]byte("Hello, world!"))
		})

	zeal.Route(router, "GET /the_answer",
		func(w zeal.ResponseWriter[int], r *zeal.Request[any]) {
			w.JSON(42)
		})

	type GetMenu struct {
		MenuID int
		Quiet  bool
	}
	zeal.Route(router, "GET /menu/{MenuID}",
		func(w zeal.ResponseWriter[models.Menu], r *zeal.Request[GetMenu]) {
			for _, menu := range menus {
				if menu.ID == r.Params.MenuID {
					w.JSON(menu, http.StatusOK)
					if !r.Params.Quiet {
						fmt.Println("Found menu: ", menu)
						fmt.Println("Returning early")
					}
					return
				}
			}

			if !r.Params.Quiet {
				fmt.Println("Menu not found")
			}
			w.WriteHeader(http.StatusNotFound)
		})

	type PostItem struct {
		MenuID int
	}
	zeal.BodyRoute(router, "POST /item",
		func(w zeal.ResponseWriter[models.Item], r *zeal.Request[PostItem], body models.Item) {
			newItem := body
			if newItem.Price < 10 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			for i := range menus {
				if menus[i].ID != r.Params.MenuID {
					continue
				}

				menus[i].Items = append(menus[i].Items, newItem)
				w.JSON(newItem, http.StatusCreated)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		})

	zeal.BodyRoute(router, "PUT /item",
		func(w zeal.ResponseWriter[models.Item], r *zeal.Request[any], b models.Item) {
			updatedItem := b
			for i := range menus {
				for j := range menus[i].Items {
					if menus[i].Items[j].Name == updatedItem.Name {
						menus[i].Items[j].Price = updatedItem.Price
						w.JSON(menus[i].Items[j])
						return
					}
				}
			}

			w.WriteHeader(http.StatusNotFound)
		})

	zeal.Route(router, "DELETE /item", handleDeleteItem)
}

type DeleteItem struct {
	Name string
}

func handleDeleteItem(w zeal.ResponseWriter[any], r *zeal.Request[DeleteItem]) {
	for i := range menus {
		for j := range menus[i].Items {
			if menus[i].Items[j].Name == r.Params.Name {
				menus[i].Items = append(menus[i].Items[:i], menus[i].Items[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
