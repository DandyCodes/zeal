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
var mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

func main() {
    addRoutes(mux)

    specOptions := zeal.SpecOptions{
        ServeMux:      mux,
        Version:       "v0.1.0",
        Description:   "Example API description.",
        StripPkgPaths: []string{"main", "models", "github.com/DandyCodes/zeal"},
    }
    openAPISpec, err := zeal.NewOpenAPISpec(specOptions)
    if err != nil {
        log.Fatalf("Failed to create OpenAPI spec: %v", err)
    }

    port := 3975
    swaggerPattern := "/swagger-ui/"
    fmt.Printf("Visit http://localhost:%v%v to see API definitions\n", port, swaggerPattern)
    zeal.ServeSwaggerUI(mux, openAPISpec, "GET "+swaggerPattern)

    fmt.Printf("Listening on port %v...\n", port)
    http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
```

## Routes

Create your route by calling ***zeal.NewRoute***, passing it a ***zeal.ServeMux***:

```go
var postRoot = zeal.NewRoute[zeal.Route](mux)
```

Passing the basic ***zeal.Route*** as a type parameter to ***zeal.NewRoute*** means this route has no:

* Response type
* URL parameters
* Request body

Now, define your handler function using the newly create route:

```go
postRoot.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Hello, world!")
})
```

Routes handled by Zeal are automatically documented in the OpenAPI spec.

## Responses

Create a route definition struct and embed ***zeal.Route*** and ***zeal.HasResponse***.

This route will respond with an integer, so ***int*** is passed to ***zeal.HasResponse*** as a type parameter:

```go
type GetAnswer struct {
    zeal.Route
    zeal.HasResponse[int]
}
```

Create your route, passing your route definition as a type parameter, and define your handler function:

```go
var getAnswer = zeal.NewRoute[GetAnswer](mux)
getAnswer.HandleFunc("GET /answer", func(w http.ResponseWriter, r *http.Request) {
    getAnswer.Response(42)
})
```

The ***Response*** method will only accept data of the declared response type.

---

Type parameters passed to ***zeal.HasResponse*** can be more complex.

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
    zeal.Route
    zeal.HasResponse[[]models.Menu]
}
var getMenus = zeal.NewRoute[GetMenus](mux)
getMenus.HandleFunc("GET /menus", func(w http.ResponseWriter, r *http.Request) {
    getMenus.Response(menus)
})
```

## URL Parameters

Create a route definition struct and embed ***zeal.Route*** and ***zeal.HasParams***.

You can pass ***zeal.HasParams*** an anonymous in-line struct definition as a type parameter.

Create your route, passing your route definition as a type parameter, and define your handler function:

```go
type DeleteMenu struct {
    zeal.Route
    zeal.HasParams[struct {
        ID    int
        Quiet bool
    }]
}
var deleteMenu = zeal.NewRoute[DeleteMenu](mux)
deleteMenu.HandleFunc("DELETE /menus/{ID}", func(w http.ResponseWriter, r *http.Request) {
    if !deleteMenu.Params().Quiet {
        fmt.Println("Deleting menu")
    }

    for i := 0; i < len(menus); i++ {
        if menus[i].ID == deleteMenu.Params().ID {
            menus = append(menus[:i], menus[i+1:]...)
            w.WriteHeader(http.StatusNoContent)
            return
        }
    }

    w.WriteHeader(http.StatusNotFound)
})
```

Params found in the URL pattern (for example, 'ID' in '/menus/{ID}') will be defined as path params whilst others will be query params.

Params are converted to their declared type. If this fails, http.StatusUnprocessableEntity 422 is sent immediately.

Struct fields must be capitalized to be accessed in the route - for example, 'Quiet'.

## Request Bodies

Create a route definition struct and embed ***zeal.Route*** and ***zeal.HasBody***.

Pass the body type to ***zeal.HasBody*** as a type parameter.

Create your route, passing your route definition as a type parameter, and define your handler function:

```go
type PutItem struct {
    zeal.Route
    zeal.HasBody[models.Item]
}
var putItem = zeal.NewRoute[PutItem](mux)
putItem.HandleFunc("PUT /items", func(w http.ResponseWriter, r *http.Request) {
    item := putItem.Body()
    if item.Price < 0 {
        http.Error(w, "Price cannot be negative", http.StatusBadRequest)
        return
    }

    for i := range menus {
        for j := range menus[i].Items {
            if menus[i].Items[j].Name == item.Name {
                menus[i].Items[j].Price = item.Price
                return
            }
        }
    }

    menus[0].Items = append(menus[0].Items, item)
    w.WriteHeader(http.StatusCreated)
})
```

The body is converted to its declared type. If this fails, http.StatusUnprocessableEntity 422 is sent immediately.

Struct fields must be capitalized to be accessed in the route - for example, 'Price'.

## Miscellaneous

Use the ***HandleFuncErr*** method to create a handler function which returns an error.

Route handler functions can be defined in an outer scope:

```go
var mux = zeal.NewServeMux(http.NewServeMux(), "Example API")

type PostItem struct {
    zeal.Route
    zeal.HasParams[struct{ MenuID int }]
    zeal.HasBody[models.Item]
    zeal.HasResponse[models.Item]
}

var postItem = zeal.NewRoute[PostItem](mux)

func addOuterScopeRoute() {
    postItem.HandleFuncErr("POST /items/{MenuID}", HandlePostItem)
}

func HandlePostItem(w http.ResponseWriter, r *http.Request) error {
    item := postItem.Body()
    if item.Price < 0 {
        return zeal.Error(w, "Price cannot be negative", http.StatusBadRequest)
    }

    for i := range menus {
        if menus[i].ID == postItem.Params().MenuID {
            menus[i].Items = append(menus[i].Items, item)
            return postItem.Response(item, http.StatusCreated)
        }
    }

    return zeal.WriteHeader(w, http.StatusNotFound)
}
```

The ***zeal.Error*** function returns a nil error after calling ***http.Error*** with a given error message and HTTP status code.

The ***Response*** method can be passed an optional HTTP status code (200 OK is sent by default). It returns a nil error if successful. Otherwise, it returns the JSON serialization error after calling ***http.Error*** with ***http.StatusInternalServerError***.

The ***zeal.WriteHeader*** function returns a nil error after calling ***http.ResponseWriter.WriteHeader*** with a given HTTP status code.

## Credits

<a href="https://www.flaticon.com/free-icons/helmet" title="helmet icons">Helmet icons created by Freepik - Flaticon</a>
