# Zeal

A type-safe REST API framework for Go!

Inspired by [FastAPI](https://github.com/tiangolo/fastapi), types are used to define and validate URL parameters, request bodies and responses.

Automatically generates OpenAPI 3 schema documentation and serves it using Swagger.

It builds upon [chi](https://github.com/go-chi/chi) for routing and [REST](https://github.com/a-h/rest) for generating the OpenAPI spec.

---

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

func main() {
  rt := zeal.NewRouter("API")

  addRoutes(rt)

  rt.Api.StripPkgPaths = []string{"main", "models", "github.com/DandyCodes/zeal"}
  spec := rt.CreateSpec("v1.0.0", "Spec")
  rt.ServeSwaggerUI(spec, "/swagger-ui*")

  fmt.Println("Listening on port 3000...")
  fmt.Println("Visit http://localhost:3000/swagger-ui to see API definitions")
  http.ListenAndServe(":3000", rt)
}

func addRoutes(rt *zeal.Router) {
  // A zeal route
  zeal.Get(rt, "/", func(w zeal.Writer[any], r *zeal.Rqr[any]) {
    w.Write([]byte("Hello, world!"))
  })

  // This route responds with an integer. Instead of no type (any),
  // the response type (int) is passed to the writer - zeal.Writer[int]
  zeal.Get(rt, "/the_answer", func(w zeal.Writer[int], r *zeal.Rqr[any]) {
    // This JSON convenience method will only accept data of the declared response type
    w.JSON(42, http.StatusOK)
  })

  // This route responds with a slice of menus - []models.Menu
  // The response type is passed to the writer - zeal.Writer[[]models.Menu]
  zeal.Get(rt, "/menus", func(w zeal.Writer[[]models.Menu], r *zeal.Rqr[any]) {
    // http status is optional and can be omitted (will send http.StatusOK 200 by default)
    w.JSON(menus)
  })

  // Struct type definition representing both path and query URL parameters
  // Fields must be capitalized so that Go exports them
  // Parameters are automatically validated and converted to their declared type
  // If validation fails, zeal responds with http.StatusUnprocessableEntity 422
  type GetPrintParams struct {
    SP   string
    IP   int
    BQ   bool
    F32Q float32
  }
  // The parameters type is passed to the request - *zeal.Rqr[GetPrintParams]
  // This route has no response type so any is passed to the writer - zeal.Writer[any]
  zeal.Get(rt, "/print/{IP}/{SP}", func(w zeal.Writer[any], r *zeal.Rqr[GetPrintParams]) {
    // IP and SP are path params because they appear in the URL path "/print/{IP}/{SP}"
    // BQ and F32Q are, therefore, automatically query params
    // Both kinds of validated params are found in the Params field of the request

    aStringPathParameter := r.Params.SP
    fmt.Println(aStringPathParameter, reflect.TypeOf(aStringPathParameter)) // string type

    anIntPathParameter := r.Params.IP
    fmt.Println(anIntPathParameter, reflect.TypeOf(anIntPathParameter)) // int type

    aBooleanQueryParameter := r.Params.BQ
    fmt.Println(aBooleanQueryParameter, reflect.TypeOf(aBooleanQueryParameter)) // bool type

    aFloat32QueryParameter := r.Params.F32Q
    fmt.Println(aFloat32QueryParameter, reflect.TypeOf(aFloat32QueryParameter)) // float32 type

    // If a query param and path param sharing the same name are sent in a single request,
    // the path param takes precedence - the value will be that of the path param
  })

  // Parameter type passed to the request - *zeal.Rqr[GetMenuParams]
  type GetMenuParams struct {
    MenuID int
  }
  // Response type passed to the writer - zeal.Writer[models.Menu]
  zeal.Get(rt, "/menu/{MenuID}", func(w zeal.Writer[models.Menu], r *zeal.Rqr[GetMenuParams]) {
    for _, menu := range menus {
      if menu.ID == r.Params.MenuID {
        w.JSON(menu)
        return
      }
    }

    w.WriteHeader(http.StatusNotFound)
  })

  // Read requests such as GET can contain parameters but never a request body
  // Write requests such as POST can contain both parameters and a request body
  type Menu struct {
    MenuID int
  }
  // Params and body types are passed to the write request - *zeal.Rqw[Menu, models.Item]
  // Request bodies are automatically validated
  // If validation fails, zeal responds with http.StatusUnprocessableEntity 422
  zeal.Post(rt, "/item", func(w zeal.Writer[models.Item], r *zeal.Rqw[Menu, models.Item]) {
    // The validated body is found in the Body field of the request
    newItem := r.Body
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

  // PUT is also a write request
  zeal.Put(rt, "/item", func(w zeal.Writer[models.Item], r *zeal.Rqw[any, models.Item]) {
    updatedItem := r.Body
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

  // DELETE is also a write request
  // Params type and handler function declared in outer scope
  zeal.Delete(rt, "/item", handleDeleteItem)
}

type DeleteItemParams struct {
  ItemName string
}

func handleDeleteItem(w zeal.Writer[any], r *zeal.Rqw[DeleteItemParams, any]) {
  for i := range menus {
    for j := range menus[i].Items {
      if menus[i].Items[j].Name == r.Params.ItemName {
        menus[i].Items = append(menus[i].Items[:i], menus[i].Items[i+1:]...)
        w.WriteHeader(http.StatusNoContent)
        return
      }
    }
  }

  w.WriteHeader(http.StatusNotFound)
}
```
