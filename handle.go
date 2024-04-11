package zeal

import (
	"net/http"
	"reflect"
	"strings"
)

func (mux *Route) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	routeValue := defineRoute(mux, pattern)
	wrapped := wrapHandlerFunc(routeValue, handlerFunc)
	mux.ZealMux.HandleFunc(pattern, wrapped)
}

func wrapHandlerFunc(routeValue reflect.Value, handlerFunc http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := initRoute(routeValue, w, r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		handlerFunc(w, r)
	}
}

type HandlerFuncErr func(http.ResponseWriter, *http.Request) error

func (mux *Route) HandleFuncErr(pattern string, handlerFunc HandlerFuncErr) {
	routeValue := defineRoute(mux, pattern)
	wrapped := wrapHandlerFuncErr(routeValue, handlerFunc)
	mux.ZealMux.HandleFunc(pattern, wrapped)
}

func wrapHandlerFuncErr(routeValue reflect.Value, handlerFunc HandlerFuncErr) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := initRoute(routeValue, w, r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		handlerFunc(w, r)
	}
}

func defineRoute(route *Route, pattern string) reflect.Value {
	routeValues := reflect.ValueOf(route).MethodByName("Validate").Call([]reflect.Value{})
	routeValue := routeValues[0].Elem().Elem().Elem()
	registerRoute(route.ZealMux, pattern, routeValue)
	return routeValue
}

func initRoute(routeValue reflect.Value, w http.ResponseWriter, r *http.Request) error {
	if routeValue.Kind() == reflect.Interface {
		return nil
	}

	paramsTypeName := getTypeName(HasParams[any]{})
	paramsValue := routeValue.FieldByName(paramsTypeName)
	if paramsValue.IsValid() {
		validateParams := paramsValue.Addr().MethodByName("Validate")
		paramsAndParamsErr := validateParams.Call([]reflect.Value{reflect.ValueOf(r)})
		err := paramsAndParamsErr[1].Interface()
		if err != nil {
			return err.(error)
		}
	}

	bodyTypeName := getTypeName(HasBody[any]{})
	bodyValue := routeValue.FieldByName(bodyTypeName)
	if bodyValue.IsValid() {
		validateBody := bodyValue.Addr().MethodByName("Validate")
		bodyAndBodyErr := validateBody.Call([]reflect.Value{reflect.ValueOf(r)})
		err := bodyAndBodyErr[1].Interface()
		if err != nil {
			return err.(error)
		}
	}

	responseTypeName := getTypeName(HasResponse[any]{})
	responseValue := routeValue.FieldByName(responseTypeName)
	if responseValue.IsValid() {
		validateResponse := responseValue.Addr().MethodByName("Validate")
		validateResponse.Call([]reflect.Value{reflect.ValueOf(&w)})
	}

	return nil
}

func getTypeName(instance any) string {
	t := reflect.TypeOf(instance)
	fullTypeName := t.String()

	lastDotIndex := strings.LastIndex(fullTypeName, ".")
	baseTypeName := fullTypeName[lastDotIndex+1:]
	baseTypeName = strings.Split(baseTypeName, "[")[0]

	return baseTypeName
}
