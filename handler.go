package zeal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/a-h/rest"
)

type ctx interface {
	Status(int)
}

type Ctx[T_Params any] struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Params         T_Params
}

func (c Ctx[T_Params]) Status(status int) {
	c.ResponseWriter.WriteHeader(status)
}

func ensureErrorCode(status int) int {
	codeStr := strconv.Itoa(status)
	if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
		return status
	}
	return 500
}

func Error[T_Response any](c ctx, status int) T_Response {
	var response T_Response
	repsponseType := reflect.TypeOf(response)
	newResponseStruct := reflect.New(repsponseType).Elem()
	response = newResponseStruct.Interface().(T_Response)
	c.Status(ensureErrorCode(status))
	return response
}

type pingHandler[T_Params any] func(Ctx[T_Params])

func (handler pingHandler[T_Params]) Unwrap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[T_Params](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		handler(Ctx[T_Params]{w, r, params})
	}
}

func Ping[T_Params any](router *Router, pattern string, handler pingHandler[T_Params]) {
	routeSchema := getRouteSchema[T_Params, any, any](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, handler.Unwrap())
}

type pullHandler[T_Params, T_Response any] func(Ctx[T_Params]) T_Response

func (handler pullHandler[T_Params, T_Response]) Unwrap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[T_Params](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		data := handler(Ctx[T_Params]{w, r, params})
		sendJSON(w, data)
	}
}

func Pull[T_Params, T_Response any](router *Router, pattern string, handler pullHandler[T_Params, T_Response]) {
	routeSchema := getRouteSchema[T_Params, any, T_Response](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, handler.Unwrap())
}

type pushHandler[T_Params, T_Body any] func(Ctx[T_Params], T_Body)

func (handler pushHandler[T_Params, T_Body]) Unwrap() http.HandlerFunc {
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

		handler(Ctx[T_Params]{w, r, params}, body)
	}
}

func Push[T_Params, T_Body any](router *Router, pattern string, handler pushHandler[T_Params, T_Body]) {
	routeSchema := getRouteSchema[T_Params, T_Body, any](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, handler.Unwrap())
}

type tradeHandler[T_Params, T_Body, T_Response any] func(Ctx[T_Params], T_Body) T_Response

func (handler tradeHandler[T_Params, T_Body, T_Response]) Unwrap() http.HandlerFunc {
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

		data := handler(Ctx[T_Params]{w, r, params}, body)
		sendJSON(w, data)
	}
}

func Trade[T_Params, T_Body, T_Response any](router *Router, pattern string, handler tradeHandler[T_Params, T_Body, T_Response]) {
	routeSchema := getRouteSchema[T_Params, T_Body, T_Response](pattern)
	registerRoute(pattern, router, routeSchema)
	router.HandleFunc(pattern, handler.Unwrap())
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

func sendJSON(w http.ResponseWriter, data any) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}
	json.NewEncoder(w).Encode(data)
}
