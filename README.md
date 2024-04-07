<div align="center">
    <div>&nbsp; &nbsp; &nbsp;<img width="150" height="150" src="./knight.png" alt="Logo"></div>
    <h1 align="center"><b>Zeal</b></h1>
    <p align="center">✨ A type-safe REST API framework for Go!</p>
</div>

## About

* Uses the standard library http.HandlerFunc for maximum compatibility.

* Define structs to validate URL parameters, request bodies and responses.

* URL parameters and request bodies are automatically converted to their declared type.

* Automatically generates fully typed OpenAPI 3 spec documentation using [REST](https://github.com/a-h/rest) and serves it with SwaggerUI.

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

A standard library http.HandlerFunc is passed to a zeal route.

Create a route definition struct and pass it to ***zeal.Route*** as a type parameter:

```go
type PostRoot struct{}
var postRoot = zeal.Route[PostRoot](Mux)
postRoot.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Hello, world!")
    w.WriteHeader(http.StatusOK)
})
```

This route uses an empty struct which means it has no:

* Response type
* URL parameters
* Request body

Routes handled by Zeal are automatically documented in the OpenAPI spec.

Using ***any*** in place of an empty struct accomplishes the same outcome:

```go
zeal.Route[any](Mux).HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Hello, world!")
})
```

## Responses

Embed ***zeal.RouteResponse*** in your route definition, passing it the response type as a type parameter.

This route responds with an ***int***:

```go
type GetAnswer struct{ zeal.RouteResponse[int] }
var getAnswer = zeal.Route[GetAnswer](Mux)
getAnswer.HandleFunc("GET /answer", func(w http.ResponseWriter, r *http.Request) {
    getAnswer.Route.Response(42)
})
```

The ***Route.Response*** method will only accept data of the declared response type.

---

A ***zeal.RouteResponse*** can use complex types.

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

This route responds with a slice of menus:

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

Embed ***zeal.RouteParams*** in your route definition, passing it the params type as a type parameter.

The params struct can be anonymous and defined in-line:

```go
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
        fmt.Println("Getting menu")
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
```

Params found in the URL pattern (for example, 'ID' in '/menus/{ID}') will be defined as path params - all others will be query params.

Params are converted to their declared type. If this fails, http.StatusUnprocessableEntity 422 is sent immediately.

Struct fields must be capitalized to be accessed in the route - for example, 'Quiet'.

## Request Bodies

Embed ***zeal.RouteBody*** in your route definition, passing it the body type as a type parameter:

```go
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
```

The body is converted to its declared type. If this fails, http.StatusUnprocessableEntity 422 is sent immediately.

Struct fields must be capitalized to be accessed in the route - for example, 'Price'.

## Miscellaneous

Use the ***HandleErr*** method to create a handler function which returns an error.

Route handler functions can be defined in an outer scope:

```go
var Mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

type PostItem struct {
    zeal.RouteParams[struct{ MenuID int }]
    zeal.RouteBody[models.Item]
}

var postItem = zeal.Route[PostItem](Mux)

func addOuterScopeRoute() {
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
```

The ***Status*** method responds with a given HTTP status code.

The ***Error*** method responds with a given HTTP status code and an optional error message. It must be passed an error code (4xx or 5xx), or else it will respond with http.StatusInternalServerError 500 instead.

## Credits

<a href="https://www.flaticon.com/free-icons/helmet" title="helmet icons">Helmet icons created by Freepik - Flaticon</a>
