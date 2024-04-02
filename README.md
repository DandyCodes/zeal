# Zeal

A type-safe REST API framework for Go!

Structs can be used to define and validate URL parameters, request bodies and response types.

Params and request bodies are automatically converted to their declared types for easy use within your routes.

Automatically generates fully typed OpenAPI 3 schema documentation using [REST](https://github.com/a-h/rest) and serves it with SwaggerUI.

## Server

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

## Routes

Routes handled by Zeal are automatically documented in the OpenAPI schema

This route has no response type, no URL params and no request body

```go
zeal.Handle(mux, "POST /",
    func(response zeal.Response[any], params any, body any) error {
        fmt.Println("Hello, world!")
        return response.Status(http.StatusOK)
    })
```

The **Status** method responds with a given HTTP status code

## Responses

The response type is passed as a type parameter

This route responds with an int - zeal.Response[int]

```go
zeal.Handle(mux, "GET /the_answer",
    func(r zeal.Response[int], p any, b any) error {
        return r.JSON(42)
    })
```

The **JSON** method will only accept data of the declared response type

Here is some example data

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

This route responds with a slice of menus - zeal.Response[[]models.Menu]

```go
var menus = []models.Menu{foodMenu, drinksMenu}

zeal.Handle(mux, "GET /menus",
    func(r zeal.Response[[]models.Menu], p any, b any) error {
        return r.JSON(menus, http.StatusOK)
    })
```

The **JSON** method can be passed an optional HTTP status code (200 OK is sent by default)

## Params

Params can be query or path params

Struct fields must begin with a capital letter

Struct representing URL params can be defined in-line

```go
zeal.Handle(mux, "GET /menus/{ID}",
    func(r zeal.Response[models.Menu], p struct{ ID int }, b any) error {
        for _, menu := range menus {
            if menu.ID == p.ID {
                return r.JSON(menu)
            }
        }

        return r.Error(http.StatusNotFound)
    })
```

Params are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

## Bodies

Request bodies are converted to their declared type

If this fails, http.StatusUnprocessableEntity 422 is sent immediately

Struct fields must be capitalized

```go
type PutItemsParams struct {
    Quiet bool
}
zeal.Handle(mux, "PUT /items",
    func(r zeal.Response[models.Item], p PutItemsParams, item models.Item) error {
        if item.Price < 0 {
            return r.Error(http.StatusBadRequest, "Price cannot be negative")
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
                return r.JSON(updatedItem)
            }
        }

        if !p.Quiet {
            fmt.Println("Creating new item:", item)
        }
        menus[1].Items = append(menus[1].Items, item)
        updatedItem := menus[1].Items[len(menus[1].Items)-1]
        return r.JSON(updatedItem, http.StatusCreated)
    })
```

## Errors

The **Error** method takes an HTTP status code and an optional error message

```go
zeal.Handle(mux, "POST /items", HandlePostItem)

func HandlePostItem(r zeal.Response[any], p struct{ MenuID int }, item models.Item) error {
    if item.Price < 0 {
        return r.Error(http.StatusBadRequest, "Price cannot be negative")
    }

    for i := range menus {
        if menus[i].ID != p.MenuID {
            continue
        }

        menus[i].Items = append(menus[i].Items, item)
        return r.Status(http.StatusCreated)
    }

    return r.Error(http.StatusNotFound)
}
```

The **Error** method must be passed a 4xx or 5xx (error) HTTP status code

If it is not passed an error code, it will respond with http.StatusInternalServerError 500 instead

## Standard Library Integration

The standard library *http.Request and http.ResponseWriter can still be accessed in a zeal route

```go
zeal.Handle(mux, "GET /",
    func(r zeal.Response[any], p any, b any) error {
        fmt.Println(r.Request)        // *http.Request
        fmt.Println(r.ResponseWriter) // http.ResponseWriter
        return nil
    })
```

And you can use *zeal.ServeMux to define a regular http.HandlerFunc

```go
func DefineStandardRoute(mux *zeal.ServeMux) {
    mux.HandleFunc("GET /std", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello!"))
    })
}
```

However, routes defined this way will not be documented in the OpenAPI spec

## Middleware

To add middleware such as logging, you can create your own wrapper around zeal.Handle

```go
func MyHandle[R, P, B any](
    mux *zeal.ServeMux, urlPattern string, handlerFunc zeal.HandlerFunc[R, P, B],
) {
    myHanlderFunc := func(r zeal.Response[R], p P, b B) error {
        err := handlerFunc(r, p, b)
        if err != nil {
            log.Println(err)
        }
        return err
    }
    zeal.Handle(mux, urlPattern, myHanlderFunc)
}
```

And define your route using your wrapper

```go
MyHandle(mux, "GET /errors_logged",
    func(r zeal.Response[[]models.Menu], p any, b any) error {
        if rand.Float64() < 0.33 {
            return r.Error(http.StatusInternalServerError, "an error occurred")
        } else {
            return r.JSON(menus)
        }
    })
```

---

Or create a stack of middleware

```go
func LogErrorHandle[R, P, B any](next zeal.HandlerFunc[R, P, B]) zeal.HandlerFunc[R, P, B] {
    return func(r zeal.Response[R], p P, b B) error {
        err := next(r, p, b)
        if err != nil {
            log.Println(err)
        }
        return err
    }
}

func LogRequestHandle[R, P, B any](next zeal.HandlerFunc[R, P, B]) zeal.HandlerFunc[R, P, B] {
    return func(r zeal.Response[R], p P, b B) error {
        log.Println(r.Request)
        return next(r, p, b)
    }
}

func AntiDdosHandle[R, P, B any](next zeal.HandlerFunc[R, P, B]) zeal.HandlerFunc[R, P, B] {
    return func(r zeal.Response[R], p P, b B) error {
        if rand.Float64() < 0.33 {
            return r.Error(http.StatusTeapot, "computer says no")
        }
        return next(r, p, b)
    }
}

func MyStackHandle[R, P, B any](
    mux *zeal.ServeMux, urlPattern string, handlerFunc zeal.HandlerFunc[R, P, B],
) {
    logErrorHandlerFunc := LogErrorHandle(handlerFunc)
    logRequestHandlerFunc := LogRequestHandle(logErrorHandlerFunc)
    antiDdosHandlerFunc := AntiDdosHandle(logRequestHandlerFunc)
    zeal.Handle(mux, urlPattern, antiDdosHandlerFunc)
}
```

And define your route using this wrapper

```go
MyStackHandle(mux, "GET /stack",
    func(r zeal.Response[[]models.Menu], p any, b any) error {
        if rand.Float64() < 0.33 {
            return r.Error(http.StatusInternalServerError, "an error occurred")
        } else {
            return r.JSON(menus)
        }
    })
```
