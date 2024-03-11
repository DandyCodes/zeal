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
	zeal.Handle(r, "POST /", func(c zeal.Ctx[any], params any, body any) {
		fmt.Println("Hello, world!")
	})

	zeal.Handle(r, "GET /answer", func(c zeal.Ctx[int], p any, b any) {
		c.JSON(42)
	})

	zeal.Handle(r, "GET /menus", func(c zeal.Ctx[[]models.Menu], p any, b any) {
		c.JSON(menus, http.StatusOK)
	})

	zeal.Handle(r, "GET /menus/{ID}", func(c zeal.Ctx[models.Menu], p struct{ ID int }, b any) {
		for _, menu := range menus {
			if menu.ID == p.ID {
				c.JSON(menu)
				return
			}
		}
		c.Error(http.StatusNotFound)
	})

	type PutItemParams struct {
		Quiet bool
	}
	zeal.Handle(r, "PUT /items",
		func(c zeal.Ctx[models.Item], p PutItemParams, item models.Item) {
			if item.Price < 0 {
				c.Error(http.StatusBadRequest, "Price cannot be negative")
				return
			}

			for i := range menus {
				for j := range menus[i].Items {
					if menus[i].Items[j].Name == item.Name {
						if !p.Quiet {
							fmt.Println("Updating item:", item)
						}
						menus[i].Items[j].Price = item.Price
						updatedItem := menus[i].Items[j]
						c.JSON(updatedItem)
						return
					}
				}
			}

			if !p.Quiet {
				fmt.Println("Creating new item:", item)
			}
			menus[0].Items = append(menus[0].Items, item)
			updatedItem := menus[0].Items[len(menus[0].Items)-1]
			c.JSON(updatedItem, http.StatusCreated)
		})

	zeal.Handle(r, "POST /items", HandlePostItem)
}

func HandlePostItem(c zeal.Ctx[any], p struct{ MenuID int }, newItem models.Item) {
	if newItem.Price < 0 {
		c.Error(http.StatusBadRequest, "Price cannot be negative")
		return
	}

	for i := range menus {
		if menus[i].ID != p.MenuID {
			continue
		}

		menus[i].Items = append(menus[i].Items, newItem)
		c.Status(http.StatusCreated)
		return
	}

	c.Error(http.StatusNotFound)
}
