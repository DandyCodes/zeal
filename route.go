package zeal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

func NewRoute[T_Route http.Handler](mux *ZealMux) *T_Route {
	var route T_Route

	routePtrValue := reflect.New(reflect.TypeOf(route))
	routeInterfacePtr := reflect.New(reflect.TypeOf(new(any)).Elem())
	routeInterfacePtr.Elem().Set(routePtrValue)

	routeStructValue := routePtrValue.Elem()

	routeMux := routeStructValue.FieldByName("Route")
	if routeMux.IsValid() {
		routeMux.FieldByName("ZealMux").Set(reflect.ValueOf(mux))
		routeMux.Addr().MethodByName("Validate").Call([]reflect.Value{routeInterfacePtr})
	} else {
		routeStructValue.Addr().MethodByName("Validate").Call([]reflect.Value{routeInterfacePtr})
		zealMux := routeStructValue.FieldByName("ZealMux")
		zealMux.Set(reflect.ValueOf(mux))
	}

	return routePtrValue.Interface().(*T_Route)
}

type Route struct {
	*ZealMux
	routeDefinition *any
}

func (r *Route) Validate(routeDefinition ...*any) *any {
	if len(routeDefinition) > 0 {
		r.routeDefinition = routeDefinition[0]
	}

	return r.routeDefinition
}

type HasParams[T_Params any] struct {
	request *http.Request
}

func (p *HasParams[T_Params]) Params() T_Params {
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
			rawParamValue := p.request.PathValue(field.Name)
			if rawParamValue == "" {
				rawParamValue = p.request.URL.Query().Get(field.Name)
			}
			paramValue, _ := parsePrimitive(rawParamValue, field.Type)
			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	params = paramsValue.Interface().(T_Params)

	return params
}

func (p *HasParams[T_Params]) Validate(request *http.Request) (T_Params, error) {
	p.request = request

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
			rawParamValue := p.request.PathValue(field.Name)
			if rawParamValue == "" {
				rawParamValue = p.request.URL.Query().Get(field.Name)
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
	request *http.Request
}

func (b *HasBody[T_Body]) Body() T_Body {
	var body T_Body
	defer b.request.Body.Close()
	json.NewDecoder(b.request.Body).Decode(&body)
	return body
}

func (b *HasBody[T_Body]) Validate(request *http.Request) (T_Body, error) {
	b.request = request

	var body T_Body
	bodyType := reflect.TypeOf(body)
	if bodyType == nil {
		return body, nil
	}

	defer b.request.Body.Close()
	bodyBytes, err := io.ReadAll(b.request.Body)
	if err != nil {
		return body, nil
	}

	// Replace the original body with a new reader based on the read bytes
	b.request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	decoder.DisallowUnknownFields() // Enable strict mode

	if err := decoder.Decode(&body); err != nil {
		return body, err
	}

	return body, nil
}

type HasResponse[T_Response any] struct {
	responseWriter *http.ResponseWriter
}

func (r *HasResponse[T_Response]) Response(data T_Response, status ...int) error {
	(*r.responseWriter).Header().Add("Content-Type", "application/json")

	if len(status) > 0 {
		(*r.responseWriter).WriteHeader(status[0])
	}

	if err := json.NewEncoder((*r.responseWriter)).Encode(data); err != nil {
		http.Error((*r.responseWriter), http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	return nil
}

func (r *HasResponse[T_Response]) Validate(responseWriter *http.ResponseWriter) {
	r.responseWriter = responseWriter
}

func WriteHeader(w http.ResponseWriter, statusCode int) error {
	w.WriteHeader(statusCode)
	return nil
}

func Error(w http.ResponseWriter, error string, code int) error {
	http.Error(w, error, code)
	return nil
}
