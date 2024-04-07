<div align="center">
    <div>&nbsp; &nbsp; &nbsp;<img width="150" height="150" src="./knight.png" alt="Logo"></div>
    <h1 align="center"><b>Zeal</b></h1>
    <p align="center">✨ A type-safe REST API framework for Go!</p>
</div>

## About

Use structs to define and validate URL parameters, request bodies and responses.

URL parameters and request bodies are automatically converted to their declared type.

Automatically generates fully typed OpenAPI 3 schema documentation using [REST](https://github.com/a-h/rest) and serves it with SwaggerUI.

## Server

```go
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
```

## Routes

Routes handled by Zeal are automatically documented in the OpenAPI schema.

Use ***zeal.Route*** to create a route and define a standard library http.HandlerFunc:

```go
zeal.Route[any](Mux).HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Hello, world!")
    w.WriteHeader(http.StatusOK)
})
```

This route uses ***any*** as a type parameter for ***zeal.Route\[any\](Mux)*** which means the route has no:

* URL parameters
* Request body
* Response type

## Responses

Define a route schema struct and embed a ***zeal.RouteResponse***:

```go
type GetAnswer struct {
    zeal.RouteResponse[int]
}
```

This route will respond with an integer, so ***int*** is passed as a type parameter.

Create the route and pass ***GetAnswer*** as a type parameter to ***zeal.Route***:

```go
var getAnswer = zeal.Route[GetAnswer](Mux)
```

Then use the ***getAnswer*** route to create the handler function:

```go
getAnswer.HandleFunc("GET /answer", func(w http.ResponseWriter, r *http.Request) {
    getAnswer.Route.Response(42)
})
```

The ***Route.Response*** method will only accept data of the declared response type.

---

You can also define complex response types.

Here is some example data:

```go
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
```

The route below responds with a slice of menus, so ***[]models.Menu*** is passed to ***zeal.RouteResponse***:

```go
var menus = []models.Menu{foodMenu, drinksMenu}

type GetMenus struct {
    zeal.RouteResponse[[]models.Menu]
}
var getMenus = zeal.Route[GetMenus](Mux)
getMenus.HandleFunc("GET /menus", func(w http.ResponseWriter, r *http.Request) {
    getMenus.Route.Response(menus, http.StatusOK)
})
```

The ***Route.Response*** method can be passed an optional HTTP status code (200 OK is sent by default).

## URL Parameters

Embed ***zeal.RouteQuery*** in your route definition to define query parameters.

Embed ***zeal.RoutePath*** to define path parameters.

Parameter struct fields must begin with a capital letter to be accessed in the route - for example, 'Quiet':

```go
type GetMenu struct {
    zeal.RouteQuery[struct{ Quiet bool }]
    zeal.RoutePath[struct{ ID int }]
    zeal.RouteResponse[models.Menu]
}
var getMenu = zeal.Route[GetMenu](Mux)
getMenu.HandleFunc("GET /menus/{ID}", func(w http.ResponseWriter, r *http.Request) {
    quiet := getMenu.Route.Query().Quiet
    if !quiet {
        fmt.Println("Getting menus")
    }

    ID := getMenu.Route.Path().ID
    for i := 0; i < len(menus); i++ {
        menu := menus[i]
        if menu.ID == ID {
            getMenu.Route.Response(menu)
            return
        }
    }

    getMenu.Error(http.StatusNotFound)
})
```

Parameters are converted to their declared type.

If this fails, http.StatusUnprocessableEntity 422 is sent immediately.

## Error Handling

Use the ***Handle*** method to create a handler function which returns an error:

```go
type GetMenu struct {
    zeal.RouteQuery[struct{ Quiet bool }]
    zeal.RoutePath[struct{ ID int }]
    zeal.RouteResponse[models.Menu]
}
var getMenu = zeal.Route[GetMenu](Mux)
getMenu.Handle("GET /menus/err/{ID}", func(w http.ResponseWriter, r *http.Request) error {
    quiet := getMenu.Route.Query().Quiet
    if !quiet {
        fmt.Println("Getting menus")
    }

    ID := getMenu.Route.Path().ID
    for i := 0; i < len(menus); i++ {
        menu := menus[i]
        if menu.ID == ID {
            return getMenu.Route.Response(menu)
        }
    }

    return getMenu.Error(http.StatusNotFound)
})
```

## Request Bodies

Embed ***zeal.RouteBody*** to define a request body:

```go
type PutItem struct {
    zeal.RouteBody[models.Item]
    zeal.RouteResponse[models.Item]
}
var putItem = zeal.Route[PutItem](Mux)
putItem.Handle("PUT /items", func(w http.ResponseWriter, r *http.Request) error {
    item := putItem.Route.Body()
    if item.Price < 0 {
        return putItem.Error(http.StatusBadRequest, "Price cannot be negative")
    }

    for i := range menus {
        for j := range menus[i].Items {
            if menus[i].Items[j].Name != item.Name {
                continue
            }

            menus[i].Items[j].Price = item.Price
            updatedItem := menus[i].Items[j]
            return putItem.Route.Response(updatedItem)
        }
    }

    menus[1].Items = append(menus[1].Items, item)
    updatedItem := menus[1].Items[len(menus[1].Items)-1]
    return putItem.Route.Response(updatedItem, http.StatusCreated)
})
```

If the body cannot be converted to its declared type, http.StatusUnprocessableEntity 422 is sent immediately.

Body struct fields must be capitalized to be accessed in the route.

## Misc

Route handler functions can be defined in an outer scope:

```go
var Mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

type PostItem struct {
    zeal.RoutePath[struct{ MenuID int }]
    zeal.RouteBody[models.Item]
}

var postItem = zeal.Route[PostItem](Mux)

func addOuterScopeRoutes() {
    postItem.Handle("POST /items/{MenuID}", HandlePostItem)
}

func HandlePostItem(w http.ResponseWriter, r *http.Request) error {
    item := postItem.Route.Body()
    if item.Price < 0 {
        return postItem.Error(http.StatusBadRequest, "Price cannot be negative")
    }

    MenuID := postItem.Route.Path().MenuID
    for i := range menus {
        if menus[i].ID != MenuID {
            continue
        }

        menus[i].Items = append(menus[i].Items, item)
        return postItem.Status(http.StatusCreated)
    }

    return postItem.Error(http.StatusNotFound)
}
```

The ***Status*** method responds with a given HTTP status code.

The ***Error*** method responds with a given HTTP status code and an optional error message. It must be passed an error code (4xx or 5xx), or else it will respond with http.StatusInternalServerError 500 instead.

###### Credits

<a href="https://www.flaticon.com/free-icons/helmet" title="helmet icons">Helmet icons created by Freepik - Flaticon</a>
