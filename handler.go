package zeal

import (
	"fmt"
	"net/http"
)

type HandlerFuncRead[ResponseType, ParamsType any] func(Writer[ResponseType], *Rqr[ParamsType])
type HandlerFuncWrite[ResponseType, ParamsType, BodyType any] func(Writer[ResponseType], *Rqw[ParamsType, BodyType])

func Get[ResponseType, ParamsType any](router *Router, pattern string, handlerFunc HandlerFuncRead[ResponseType, ParamsType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, any](pattern)
	route := router.Api.Get(pattern)
	registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
	registerRequestModel(route, routeSchema.bodyType)
	registerResponseModel(route, routeSchema.responseType)
	router.Mux.Get(pattern, unwrapHandlerFuncRead(handlerFunc))
}

func Post[ResponseType, ParamsType, BodyType any](router *Router, pattern string, handlerFunc HandlerFuncWrite[ResponseType, ParamsType, BodyType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, BodyType](pattern)
	route := router.Api.Post(pattern)
	registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
	registerRequestModel(route, routeSchema.bodyType)
	registerResponseModel(route, routeSchema.responseType)
	router.Mux.Post(pattern, unwrapHandlerFuncWrite(handlerFunc))
}

func Put[ResponseType, ParamsType, BodyType any](router *Router, pattern string, handlerFunc HandlerFuncWrite[ResponseType, ParamsType, BodyType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, BodyType](pattern)
	route := router.Api.Put(pattern)
	registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
	registerRequestModel(route, routeSchema.bodyType)
	registerResponseModel(route, routeSchema.responseType)
	router.Mux.Put(pattern, unwrapHandlerFuncWrite(handlerFunc))
}

func Delete[ResponseType, ParamsType, BodyType any](router *Router, pattern string, handlerFunc HandlerFuncWrite[ResponseType, ParamsType, BodyType]) {
	routeSchema := getRouteSchema[ResponseType, ParamsType, BodyType](pattern)
	route := router.Api.Delete(pattern)
	registerParameters(route, routeSchema.pattern, routeSchema.paramsType)
	registerRequestModel(route, routeSchema.bodyType)
	registerResponseModel(route, routeSchema.responseType)
	router.Mux.Delete(pattern, unwrapHandlerFuncWrite(handlerFunc))
}

func unwrapHandlerFuncRead[ResponseType, ParamsType any](handlerFunc HandlerFuncRead[ResponseType, ParamsType]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[ParamsType](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		rqr := Rqr[ParamsType]{Request: r, Params: params}
		handlerFunc(Writer[ResponseType]{w}, &rqr)
	}
}

func unwrapHandlerFuncWrite[ResponseType, ParamsType, BodyType any](handlerFunc HandlerFuncWrite[ResponseType, ParamsType, BodyType]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getParams[ParamsType](r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		body, ok := getBody[BodyType](r)
		if !ok {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		rqw := Rqw[ParamsType, BodyType]{Rqr: Rqr[ParamsType]{Request: r, Params: params}, Body: body}
		handlerFunc(Writer[ResponseType]{w}, &rqw)
	}
}
