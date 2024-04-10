package zeal

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

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

	paramsTypeName := getTypeName(HasParams[any]{})
	paramsValue := routeRef.FieldByName(paramsTypeName)
	if paramsValue.IsValid() {
		requestValue := paramsValue.FieldByName("Request")
		if requestValue.CanSet() {
			requestValue.Set(reflect.ValueOf(r))
		}
		paramsField, _ := reflect.TypeOf(routeRef.Interface()).FieldByName(paramsTypeName)
		validateMethod, _ := paramsField.Type.MethodByName("Validate")
		paramsAndErr := validateMethod.Func.Call([]reflect.Value{paramsValue})
		err := paramsAndErr[1].Interface()
		if err != nil {
			return routeRef.Interface().(T_Route), err.(error)
		}
	}

	bodyTypeName := getTypeName(HasBody[any]{})
	bodyValue := routeRef.FieldByName(bodyTypeName)
	if bodyValue.IsValid() {
		requestValue := bodyValue.FieldByName("Request")
		if requestValue.CanSet() {
			requestValue.Set(reflect.ValueOf(r))
		}
		bodyField, _ := reflect.TypeOf(routeRef.Interface()).FieldByName(bodyTypeName)
		validateMethod, _ := bodyField.Type.MethodByName("Validate")
		bodyAndErr := validateMethod.Func.Call([]reflect.Value{bodyValue})
		err := bodyAndErr[1].Interface()
		if err != nil {
			return routeRef.Interface().(T_Route), err.(error)
		}
	}

	responseTypeName := getTypeName(HasResponse[any]{})
	responseValue := routeRef.FieldByName(responseTypeName)
	if responseValue.IsValid() {
		responseWriterValue := responseValue.FieldByName("ResponseWriter")
		if responseWriterValue.CanSet() {
			responseWriterValue.Set(reflect.ValueOf(w))
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
