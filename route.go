package zeal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type Route struct {
	*ServeMux
	RouteDefinition *any
}

func NewRoute[T_Route http.Handler](mux *ServeMux) *T_Route {
	var route T_Route

	routePtrValue := reflect.New(reflect.TypeOf(route))
	routeInterfacePtr := reflect.New(reflect.TypeOf(new(any)).Elem())
	routeInterfacePtr.Elem().Set(routePtrValue)

	routeStructValue := routePtrValue.Elem()

	routeMux := routeStructValue.FieldByName("Route")
	if routeMux.IsValid() {
		routeMux.FieldByName("ServeMux").Set(reflect.ValueOf(mux))
		routeMux.FieldByName("RouteDefinition").Set(routeInterfacePtr)
	} else {
		routeStructValue.FieldByName("RouteDefinition").Set(routeInterfacePtr)
		serveMux := routeStructValue.FieldByName("ServeMux")
		serveMux.Set(reflect.ValueOf(mux))
	}

	return routePtrValue.Interface().(*T_Route)
}

type HasParams[T_Params any] struct {
	Request *http.Request
}

func (p HasParams[T_Params]) Params() T_Params {
	var params T_Params
	paramsType := reflect.TypeOf(params)
	if paramsType == nil {
		return params
	}

	paramsValue := reflect.New(paramsType).Elem()

	for i := 0; i < paramsType.NumField(); i++ {
		field := paramsType.Field(i)
		structField := paramsValue.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := p.Request.PathValue(field.Name)
			if rawParamValue == "" {
				rawParamValue = p.Request.URL.Query().Get(field.Name)
			}
			paramValue, _ := parsePrimitive(rawParamValue, field.Type)
			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	params = paramsValue.Interface().(T_Params)

	return params
}

func (p HasParams[T_Params]) Validate() (T_Params, error) {
	var params T_Params
	paramsType := reflect.TypeOf(params)
	if paramsType == nil {
		return params, nil
	}

	paramsValue := reflect.New(paramsType).Elem()

	var error error

	for i := 0; i < paramsType.NumField(); i++ {
		field := paramsType.Field(i)
		structField := paramsValue.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := p.Request.PathValue(field.Name)
			if rawParamValue == "" {
				rawParamValue = p.Request.URL.Query().Get(field.Name)
			}
			paramValue, err := parsePrimitive(rawParamValue, field.Type)
			if err != nil {
				error = err
				continue
			}

			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	params = paramsValue.Interface().(T_Params)

	return params, error
}

type HasBody[T_Body any] struct {
	Request *http.Request
}

func (b HasBody[T_Body]) Body() T_Body {
	var body T_Body
	defer b.Request.Body.Close()
	json.NewDecoder(b.Request.Body).Decode(&body)
	return body
}

func (b HasBody[T_Body]) Validate() (T_Body, error) {
	var body T_Body
	bodyType := reflect.TypeOf(body)
	if bodyType == nil {
		return body, nil
	}

	defer b.Request.Body.Close()
	bodyBytes, err := io.ReadAll(b.Request.Body)
	if err != nil {
		return body, nil
	}

	// Replace the original body with a new reader based on the read bytes
	b.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	decoder.DisallowUnknownFields() // Enable strict mode

	if err := decoder.Decode(&body); err != nil {
		return body, err
	}

	return body, nil
}

type HasResponse[T_Response any] struct {
	ResponseWriter http.ResponseWriter
}

func (r HasResponse[T_Response]) Response(data T_Response, status ...int) error {
	r.ResponseWriter.Header().Add("Content-Type", "application/json")

	if len(status) > 0 {
		r.ResponseWriter.WriteHeader(status[0])
	}

	if err := json.NewEncoder(r.ResponseWriter).Encode(data); err != nil {
		http.Error(r.ResponseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	return nil
}

func WriteHeader(w http.ResponseWriter, statusCode int) error {
	w.WriteHeader(statusCode)
	return nil
}

func Error(w http.ResponseWriter, error string, code int) error {
	http.Error(w, error, code)
	return nil
}
