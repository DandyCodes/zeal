package zeal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/rest"
)

type ReadHandler[ResponseType, ParamsType any] func(ResponseWriter[ResponseType], *Request[ParamsType])

func Read[ResponseType, ParamsType any](router *Router, pattern string, handler ReadHandler[ResponseType, ParamsType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, any](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, unwrapReadHandler(handler))
}

func unwrapReadHandler[T_Response, T_Params any](handler func(ResponseWriter[T_Response], *Request[T_Params])) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[T_Params](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		request := Request[T_Params]{Request: r, Params: params}
		handler(ResponseWriter[T_Response]{w}, &request)
	}
}

type WriteHandler[ResponseType, ParamsType, BodyType any] func(ResponseWriter[ResponseType], *Request[ParamsType], BodyType)

func Write[ResponseType, ParamsType, BodyType any](router *Router, pattern string, handler WriteHandler[ResponseType, ParamsType, BodyType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, BodyType](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, unwrapWriteHandler(handler))
}

func unwrapWriteHandler[T_Response, T_Params, T_Body any](handler func(ResponseWriter[T_Response], *Request[T_Params], T_Body)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[T_Params](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		body, ok := getBody[T_Body](r)
		if !ok {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		request := Request[T_Params]{Request: r, Params: params}
		handler(ResponseWriter[T_Response]{w}, &request, body)
	}
}

const (
	getPrefix    = "GET "
	postPrefix   = "POST "
	putPrefix    = "PUT "
	deletePrefix = "DELETE "
)

func registerRoute(pattern string, router *Router, routeSchema RouteSchema) {
	var route *rest.Route

	switch {
	case strings.HasPrefix(pattern, getPrefix):
		path := strings.TrimPrefix(pattern, getPrefix)
		route = router.Api.Get(path)
	case strings.HasPrefix(pattern, postPrefix):
		path := strings.TrimPrefix(pattern, postPrefix)
		route = router.Api.Post(path)
	case strings.HasPrefix(pattern, putPrefix):
		path := strings.TrimPrefix(pattern, putPrefix)
		route = router.Api.Put(path)
	case strings.HasPrefix(pattern, deletePrefix):
		path := strings.TrimPrefix(pattern, deletePrefix)
		route = router.Api.Delete(path)
	}

	if route != nil {
		registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
		registerRequestModel(route, routeSchema.bodyType)
		registerResponseModel(route, routeSchema.responseType)
	}
}
