package main

import (
	"fmt"
	"net/http"

	"github.com/DandyCodes/zeal"
	"github.com/DandyCodes/zeal/example/models"
)

func main() {
	r := zeal.NewRouter("API")

	addRoutes(r)

	r.Api.StripPkgPaths = []string{"main", "models", "github.com/DandyCodes/zeal"}
	spec := r.CreateSpec("v1.0.0", "Spec")
	r.ServeSwaggerUI(spec, "GET /swagger-ui/")

	fmt.Println("Listening on port 3000...")
	fmt.Println("Visit http://localhost:3000/swagger-ui to see API definitions")
	http.ListenAndServe(":3000", r)
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

func addRoutes(r *zeal.Router) {
	zeal.Ping(r, "POST /", func(c zeal.Ctx[any]) {
		fmt.Println("Hello, world!")
	})

	zeal.Pull(r, "GET /", func(c zeal.Ctx[any]) int {
		return 42
	})

	type GetMenu struct {
		MenuID int
	}
	zeal.Pull(r, "GET /menu/{MenuID}", func(c zeal.Ctx[GetMenu]) models.Menu {
		for _, menu := range menus {
			if menu.ID == c.Params.MenuID {
				return menu
			}
		}
		return zeal.Error[models.Menu](c, http.StatusNotFound)
	})

	zeal.Ping(r, "DELETE /item", func(c zeal.Ctx[struct{ Name string }]) {
		for i := range menus {
			for j := range menus[i].Items {
				if menus[i].Items[j].Name == c.Params.Name {
					menus[i].Items = append(menus[i].Items[:i], menus[i].Items[i+1:]...)
					c.Status(http.StatusNoContent)
					return
				}
			}
		}

		c.Status(http.StatusNotFound)
	})

	zeal.Push(r, "POST /item", func(c zeal.Ctx[struct{ MenuID int }], body models.Item) {
		newItem := body
		if newItem.Price < 0 {
			c.Status(http.StatusBadRequest)
			return
		}

		for i := range menus {
			if menus[i].ID != c.Params.MenuID {
				continue
			}
			menus[i].Items = append(menus[i].Items, newItem)
			c.Status(http.StatusCreated)
			return
		}

		c.Status(http.StatusNotFound)
	})

	zeal.Trade(r, "PUT /item", handleUpsertItem)
}

func handleUpsertItem(c zeal.Ctx[struct{ Quiet bool }], updateItem models.Item) models.Item {
	if updateItem.Price < 0 {
		return zeal.Error[models.Item](c, http.StatusBadRequest)
	}

	for i := range menus {
		for j := range menus[i].Items {
			if menus[i].Items[j].Name == updateItem.Name {
				if !c.Params.Quiet {
					fmt.Println("Updating item: ", updateItem)
				}
				menus[i].Items[j].Price = updateItem.Price
				return updateItem
			}
		}
	}

	if !c.Params.Quiet {
		fmt.Println("Creating new item: ", updateItem)
	}
	menus[0].Items = append(menus[0].Items, updateItem)
	c.Status(http.StatusCreated)
	return updateItem
}
