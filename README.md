# Zeal

A type-safe REST API framework for Go!

Structs can be used to define and validate URL parameters, request bodies and response types.

Automatically generates fully typed OpenAPI 3 schema documentation using [REST](https://github.com/a-h/rest) and serves it with SwaggerUI.

## Usage

```go
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
```

---

Routes handled by Zeal are automatically documented in the OpenAPI schema

This route has no response type, no URL params and no request body

```go
zeal.Handle(mux, "POST /", func(r zeal.Response[any], params any, body any) {
    fmt.Println("Hello, world!")
})
```

The response type is passed as a type parameter

This route responds with an int - zeal.Response[int]

```go
zeal.Handle(mux, "GET /answer", func(r zeal.Response[int], p any, b any) {
    r.JSON(42)
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

This route responds with a slice of menus - zeal.Response[[]models.Menu]

```go
zeal.Handle(mux, "GET /menus", func(r zeal.Response[[]models.Menu], p any, b any) {
    r.JSON(menus, http.StatusOK)
})
```

The JSON method can be passed an optional HTTP status code (the route responds with 200 OK by default)

---

Params can be query or path params

Struct representing URL params can be defined in-line

```go
zeal.Handle(mux, "GET /menus/{ID}", func(r zeal.Response[models.Menu], p struct{ ID int }, b any) {
    for _, menu := range menus {
        if menu.ID == p.ID {
            r.JSON(menu)
            return
        }
    }
    r.Error(http.StatusNotFound)
})
```

Params struct fields must be capitalized (i.e. 'Quiet')

```go
type PutItemParams struct {
    Quiet bool
}
zeal.Handle(mux, "PUT /items",
    func(r zeal.Response[models.Item], p PutItemParams, item models.Item) {
        if item.Price < 0 {
            r.Error(http.StatusBadRequest, "Price cannot be negative")
            return
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
                r.JSON(updatedItem)
                return
            }
        }

        if !p.Quiet {
            fmt.Println("Creating new item:", item)
        }
        menus[1].Items = append(menus[1].Items, item)
        updatedItem := menus[1].Items[len(menus[1].Items)-1]
        r.JSON(updatedItem, http.StatusCreated)
    })
```

Params and request bodies are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

```go
zeal.Handle(mux, "POST /items", HandlePostItem)

func HandlePostItem(r zeal.Response[any], p struct{ MenuID int }, newItem models.Item) {
    if newItem.Price < 0 {
        r.Error(http.StatusBadRequest, "Price cannot be negative")
        return
    }

    for i := range menus {
        if menus[i].ID != p.MenuID {
            continue
        }

        menus[i].Items = append(menus[i].Items, newItem)
        r.Status(http.StatusCreated)
        return
    }

    r.Error(http.StatusNotFound)
}
```

The **Error** method takes an HTTP status code and an optional error message
