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

Routes handled by Zeal are automatically documented in the OpenAPI schema

This route has no response body, URL params or request body

```go
zeal.Handle(r, "POST /", func(c zeal.Ctx[any], params any, body any) {
    fmt.Println("Hello, world!")
})
```

The route context receives the response type as a type parameter

This route responds with an int - zeal.Ctx[int]

```go
zeal.Handle(r, "GET /answer", func(c zeal.Ctx[int], p any, b any) {
    c.JSON(42)
})
```

The JSON method will only accept data of the declared response type

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

This route responds with a slice of menus - zeal.Ctx[[]models.Menu]

```go
zeal.Handle(r, "GET /menus", func(c zeal.Ctx[[]models.Menu], p any, b any) {
    c.JSON(menus, http.StatusOK)
})
```

The JSON method can be passed an optional HTTP status code (the route responds with 200 OK by default)

---

Params can be query or path params

Struct representing URL params can be defined in-line

```go
zeal.Handle(r, "GET /menus/{ID}", func(c zeal.Ctx[models.Menu], p struct{ ID int }, b any) {
    for _, menu := range menus {
        if menu.ID == p.ID {
            c.JSON(menu)
            return
        }
    }
    c.Error(http.StatusNotFound)
})
```

Params struct fields must be capitalized (i.e. 'Quiet')

```go
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
```

Params and request bodies are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

```go
zeal.Handle(r, "POST /items", HandlePostItem)

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
```

The **Error** method takes an HTTP status code and an optional message
