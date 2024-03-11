package zeal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/rest"
)

type Handler[T_Response, T_Params, T_Body any] func(Ctx[T_Response], T_Params, T_Body)

func unwrapHandler[T_Response, T_Params, T_Body any](handler Handler[T_Response, T_Params, T_Body]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[T_Params](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		body, ok := getBody[T_Body](r)
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		handler(Ctx[T_Response]{w, r}, params, body)
	}
}

func Handle[T_Response, T_Params, T_Body any](router *Router, pattern string, handler Handler[T_Response, T_Params, T_Body]) {
	routeSchema := getRouteSchema[T_Response, T_Params, T_Body](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, unwrapHandler(handler))
}

const (
	connectPrefix = "CONNECT "
	deletePrefix  = "DELETE "
	getPrefix     = "GET "
	headPrefix    = "HEAD "
	optionsPrefix = "OPTIONS "
	patchPrefix   = "PATCH "
	postPrefix    = "POST "
	putPrefix     = "PUT "
	tracePrefix   = "TRACE "
)

func registerRoute(pattern string, router *Router, routeSchema routeSchema) {
	var route *rest.Route

	switch {
	case strings.HasPrefix(pattern, connectPrefix):
		path := strings.TrimPrefix(pattern, connectPrefix)
		route = router.Api.Connect(path)
	case strings.HasPrefix(pattern, deletePrefix):
		path := strings.TrimPrefix(pattern, deletePrefix)
		route = router.Api.Delete(path)
	case strings.HasPrefix(pattern, getPrefix):
		path := strings.TrimPrefix(pattern, getPrefix)
		route = router.Api.Get(path)
	case strings.HasPrefix(pattern, headPrefix):
		path := strings.TrimPrefix(pattern, headPrefix)
		route = router.Api.Head(path)
	case strings.HasPrefix(pattern, optionsPrefix):
		path := strings.TrimPrefix(pattern, optionsPrefix)
		route = router.Api.Options(path)
	case strings.HasPrefix(pattern, patchPrefix):
		path := strings.TrimPrefix(pattern, patchPrefix)
		route = router.Api.Patch(path)
	case strings.HasPrefix(pattern, postPrefix):
		path := strings.TrimPrefix(pattern, postPrefix)
		route = router.Api.Post(path)
	case strings.HasPrefix(pattern, putPrefix):
		path := strings.TrimPrefix(pattern, putPrefix)
		route = router.Api.Put(path)
	case strings.HasPrefix(pattern, tracePrefix):
		path := strings.TrimPrefix(pattern, tracePrefix)
		route = router.Api.Trace(path)
	}

	if route != nil {
		registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
		registerRequestModel(route, routeSchema.bodyType)
		registerResponseModel(route, routeSchema.responseType)
	}
}
