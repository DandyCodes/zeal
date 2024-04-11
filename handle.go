package zeal

import (
	"net/http"
	"reflect"
	"strings"
)

func (mux *Route) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	routeValue := defineRoute(mux, pattern)
	wrapped := wrapHandlerFunc(routeValue, handlerFunc)
	mux.ServeMux.HandleFunc(pattern, wrapped)
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
	mux.ServeMux.HandleFunc(pattern, wrapped)
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

func defineRoute(mux *Route, pattern string) reflect.Value {
	routeDefinitionValue := reflect.ValueOf(mux).Elem().FieldByName("RouteDefinition")
	routeValue := routeDefinitionValue.Elem().Elem().Elem()
	registerRoute(mux.ServeMux, pattern, routeValue.Type())
	return routeValue
}

func initRoute(routeValue reflect.Value, w http.ResponseWriter, r *http.Request) error {
	if routeValue.Kind() == reflect.Interface {
		return nil
	}

	paramsTypeName := getTypeName(HasParams[any]{})
	paramsValue := routeValue.FieldByName(paramsTypeName)
	if paramsValue.IsValid() {
		requestValue := paramsValue.FieldByName("Request")
		if requestValue.CanSet() {
			requestValue.Set(reflect.ValueOf(r))
		}
		paramsField, _ := reflect.TypeOf(routeValue.Interface()).FieldByName(paramsTypeName)
		validateMethod, _ := paramsField.Type.MethodByName("Validate")
		paramsAndErr := validateMethod.Func.Call([]reflect.Value{paramsValue})
		err := paramsAndErr[1].Interface()
		if err != nil {
			return err.(error)
		}
	}

	bodyTypeName := getTypeName(HasBody[any]{})
	bodyValue := routeValue.FieldByName(bodyTypeName)
	if bodyValue.IsValid() {
		requestValue := bodyValue.FieldByName("Request")
		if requestValue.CanSet() {
			requestValue.Set(reflect.ValueOf(r))
		}
		bodyField, _ := reflect.TypeOf(routeValue.Interface()).FieldByName(bodyTypeName)
		validateMethod, _ := bodyField.Type.MethodByName("Validate")
		bodyAndErr := validateMethod.Func.Call([]reflect.Value{bodyValue})
		err := bodyAndErr[1].Interface()
		if err != nil {
			return err.(error)
		}
	}

	responseTypeName := getTypeName(HasResponse[any]{})
	responseValue := routeValue.FieldByName(responseTypeName)
	if responseValue.IsValid() {
		responseWriterValue := responseValue.FieldByName("ResponseWriter")
		if responseWriterValue.CanSet() {
			responseWriterValue.Set(reflect.ValueOf(w))
		}
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
