# Zeal

A type-safe REST API framework for Go!

Structs can be used to define and validate URL parameters, request bodies and response types.

Automatically generates OpenAPI 3 schema documentation using [REST](https://github.com/a-h/rest) and serves it using Swagger.

## Usage

```go
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
```

Routes are documented in the OpenAPI spec

```go
func addRoutes(router *zeal.Router) {
    zeal.Route(router, "GET /",
       func(w zeal.ResponseWriter[any], r *zeal.Request[any]) {
          w.BodyRoute([]byte("Hello, world!"))
    })
}
```

---

This route responds with an integer - zeal.ResponseWriter[int]

```go
zeal.Route(router, "GET /the_answer",
    func(w zeal.ResponseWriter[int], r *zeal.Request[any]) {
       w.JSON(42)
    })
```

This JSON convenience method will only accept data of the declared response type

---

Example API

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

Struct with fields representing both path and query URL parameters - must be capitalized

```go
type GetMenu struct {
    MenuID int  // path param
    Quiet  bool // query param
}
zeal.Route(router, "GET /menu/{MenuID}",
    func(w zeal.ResponseWriter[models.Menu], r *zeal.Request[GetMenu]) {
        for _, menu := range menus {
            if menu.ID == r.Params.MenuID {
                w.JSON(menu, http.StatusOK) // status optional - OK 200 sent by default
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
```

Params are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

---

This route has a request body

```go
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
```

The body is converted to its declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

---

Put request with no params

```go
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
```

---

Delete request with handler function declared in outer scope

```go
zeal.Route(router, "DELETE /item", handleDeleteItem)

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
```
