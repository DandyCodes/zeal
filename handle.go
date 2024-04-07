package zeal

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type RouteMux[T_Route any] struct {
	*ServeMux
	Route          T_Route
	ResponseWriter *http.ResponseWriter
}

func Route[T_Route any](mux *ServeMux) *RouteMux[T_Route] {
	var route T_Route
	return &RouteMux[T_Route]{ServeMux: mux, Route: route}
}

func (mux *RouteMux[T_Route]) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	muxType := reflect.TypeOf(mux).Elem()
	routeStructField, ok := muxType.FieldByName("Route")
	if ok {
		registerRoute(mux.ServeMux, pattern, routeStructField)

		wrapped := func(w http.ResponseWriter, r *http.Request) {
			muxValue := reflect.ValueOf(mux).Elem()
			routeRef := reflect.New(routeStructField.Type).Elem()
			route, err := newRoute[T_Route](routeRef, w, r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}

			routeField := muxValue.FieldByName("Route")
			if routeField.CanSet() {
				routeField.Set(reflect.ValueOf(route))
			}

			responseWriterField := muxValue.FieldByName("ResponseWriter")
			if responseWriterField.CanSet() {
				responseWriterField.Set(reflect.ValueOf(&w))
			}

			handlerFunc(w, r)
		}
		mux.ServeMux.HandleFunc(pattern, wrapped)
	}
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (mux *RouteMux[T_Route]) Handle(pattern string, handlerFunc HandlerFunc) {
	muxType := reflect.TypeOf(mux).Elem()
	routeStructField, ok := muxType.FieldByName("Route")
	if ok {
		registerRoute(mux.ServeMux, pattern, routeStructField)

		wrapped := func(w http.ResponseWriter, r *http.Request) {
			muxValue := reflect.ValueOf(mux).Elem()
			routeRef := reflect.New(routeStructField.Type).Elem()
			route, err := newRoute[T_Route](routeRef, w, r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}

			routeField := muxValue.FieldByName("Route")
			if routeField.CanSet() {
				routeField.Set(reflect.ValueOf(route))
			}

			responseWriterField := muxValue.FieldByName("ResponseWriter")
			if responseWriterField.CanSet() {
				responseWriterField.Set(reflect.ValueOf(&w))
			}

			handlerFunc(w, r)
		}
		mux.ServeMux.HandleFunc(pattern, wrapped)
	}
}

func (mux *RouteMux[T_Route]) Status(status int) error {
	(*mux.ResponseWriter).WriteHeader(status)
	return nil
}

func (mux *RouteMux[T_Route]) Error(status int, message ...string) error {
	status = ensureErrorCode(status)

	var msg string
	if len(message) > 0 {
		msg = message[0]
	} else {
		msg = http.StatusText(status)
	}

	http.Error(*mux.ResponseWriter, msg, status)

	return fmt.Errorf(msg)
}

func ensureErrorCode(status int) int {
	codeStr := strconv.Itoa(status)
	if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
		return status
	}

	fmt.Printf("Expected HTTP error status code. Received: %v. Returning 500 instead.\n", status)
	return 500
}

func newRoute[T_Route any](routeRef reflect.Value, w http.ResponseWriter, r *http.Request) (T_Route, error) {
	if routeRef.Kind() == reflect.Interface {
		return reflect.ValueOf(struct{}{}).Interface().(T_Route), nil
	}
	queryTypeName := getTypeName(RouteQuery[any]{})
	query := routeRef.FieldByName(queryTypeName)
	if query.IsValid() {
		request := query.FieldByName("Request")
		if request.CanSet() {
			request.Set(reflect.ValueOf(r))
		}
		queryInstance, _ := reflect.TypeOf(routeRef.Interface()).FieldByName(queryTypeName)
		validateMethod, _ := queryInstance.Type.MethodByName("Validate")
		queryAndErr := validateMethod.Func.Call([]reflect.Value{query})
		err := queryAndErr[1].Interface()
		if err != nil {
			return routeRef.Interface().(T_Route), err.(error)
		}
	}

	pathTypeName := getTypeName(RoutePath[any]{})
	path := routeRef.FieldByName(pathTypeName)
	if path.IsValid() {
		request := path.FieldByName("Request")
		if request.CanSet() {
			request.Set(reflect.ValueOf(r))
		}
		pathInstance, _ := reflect.TypeOf(routeRef.Interface()).FieldByName(pathTypeName)
		validateMethod, _ := pathInstance.Type.MethodByName("Validate")
		pathAndErr := validateMethod.Func.Call([]reflect.Value{path})
		err := pathAndErr[1].Interface()
		if err != nil {
			return routeRef.Interface().(T_Route), err.(error)
		}
	}

	bodyTypeName := getTypeName(RouteBody[any]{})
	body := routeRef.FieldByName(bodyTypeName)
	if body.IsValid() {
		request := body.FieldByName("Request")
		if request.CanSet() {
			request.Set(reflect.ValueOf(r))
		}
		bodyInstance, _ := reflect.TypeOf(routeRef.Interface()).FieldByName(bodyTypeName)
		validateMethod, _ := bodyInstance.Type.MethodByName("Validate")
		bodyAndErr := validateMethod.Func.Call([]reflect.Value{body})
		err := bodyAndErr[1].Interface()
		if err != nil {
			return routeRef.Interface().(T_Route), err.(error)
		}
	}

	responseTypeName := getTypeName(RouteResponse[any]{})
	response := routeRef.FieldByName(responseTypeName)
	if response.IsValid() {
		responseWriter := response.FieldByName("ResponseWriter")
		if responseWriter.CanSet() {
			responseWriter.Set(reflect.ValueOf(w))
		}
	}

	return routeRef.Interface().(T_Route), nil
}

func getTypeName(instance any) string {
	t := reflect.TypeOf(instance)
	fullTypeName := t.String()

	lastDotIndex := strings.LastIndex(fullTypeName, ".")
	baseTypeName := fullTypeName[lastDotIndex+1:]
	baseTypeName = strings.Split(baseTypeName, "[")[0]

	return baseTypeName
}
