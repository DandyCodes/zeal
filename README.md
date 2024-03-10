# Zeal

A type-safe REST API framework for Go!

Structs can be used to define and validate URL parameters, request bodies and response types.

Automatically generates OpenAPI 3 schema documentation using [REST](https://github.com/a-h/rest) and serves it using Swagger.

## Usage

```go
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
```

---

```go
zeal.Ping(r, "POST /", func(c zeal.Ctx[any]) {
    fmt.Println("Hello, world!")
})
```

**Ping** doesn't return a response body but **Pull** does

```go
zeal.Pull(r, "GET /", func(c zeal.Ctx[any]) int {
    return 42
})
```

---

Example data

```go
var foodMenu = models.Menu{
    ID:    1,
    Items: []models.Item{{Name: "Steak", Price: 13.95}, {Name: "Potatoes", Price: 3.95}},
}

var drinksMenu = models.Menu{
    ID:    2,
    Items: []models.Item{{Name: "Juice", Price: 1.25}, {Name: "Soda", Price: 1.75}},
}

var menus = []models.Menu{foodMenu, drinksMenu}
```

Struct representing URL parameters - fields must be capitalized (i.e. 'MenuID')

```go
type GetMenu struct {
    MenuID int
}
zeal.Pull(r, "GET /menu/{MenuID}", func(c zeal.Ctx[GetMenu]) models.Menu {
    for _, menu := range menus {
        if menu.ID == c.Param.MenuID {
            return menu
        }
    }
    return zeal.Error[models.Menu](c, http.StatusNotFound)
})
```

Returning an **Error** ends the request with empty data of the correct type

Params can be query or path params

Params struct can be defined in-line

```go
zeal.Ping(r, "DELETE /item", func(c zeal.Ctx[struct{ Name string }]) {
    for i := range menus {
        for j := range menus[i].Items {
            if menus[i].Items[j].Name == c.Param.Name {
                menus[i].Items = append(menus[i].Items[:i], menus[i].Items[i+1:]...)
                c.W.WriteHeader(http.StatusNoContent)
                return
            }
        }
    }

    c.W.WriteHeader(http.StatusNotFound)
})
```

Params are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

**Push** takes a request body

```go
zeal.Push(r, "POST /item", func(c zeal.Ctx[struct{ MenuID int }], body models.Item) {
    newItem := body
    if newItem.Price < 0 {
        c.W.WriteHeader(http.StatusBadRequest)
        return
    }

    for i := range menus {
        if menus[i].ID != c.Param.MenuID {
            continue
        }
        menus[i].Items = append(menus[i].Items, newItem)
        c.W.WriteHeader(http.StatusCreated)
        return
    }

    c.W.WriteHeader(http.StatusNotFound)
})
```

Like params, the request body is converted to its declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

**Trade** takes a request body and returns a response body

```go
zeal.Trade(r, "PUT /item", handleUpsertItem)

func handleUpsertItem(c zeal.Ctx[struct{ Quiet bool }], updateItem models.Item) models.Item {
    if updateItem.Price < 0 {
        return zeal.Error[models.Item](c, http.StatusBadRequest)
    }

    for i := range menus {
        for j := range menus[i].Items {
            if menus[i].Items[j].Name == updateItem.Name {
                if !c.Param.Quiet {
                    fmt.Println("Updating item: ", updateItem)
                }
                menus[i].Items[j].Price = updateItem.Price
                return updateItem
            }
        }
    }

    if !c.Param.Quiet {
        fmt.Println("Creating new item: ", updateItem)
    }
    menus[0].Items = append(menus[0].Items, updateItem)
    c.W.WriteHeader(http.StatusCreated)
    return updateItem
}
```
