package zeal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type RouteQuery[T_Query any] struct {
	Request *http.Request
}

func (q RouteQuery[T_Query]) Query() T_Query {
	var query T_Query
	queryType := reflect.TypeOf(query)
	if queryType == nil {
		return query
	}

	newQueryStruct := reflect.New(queryType).Elem()

	for i := 0; i < queryType.NumField(); i++ {
		field := queryType.Field(i)
		structField := newQueryStruct.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := q.Request.URL.Query().Get(field.Name)
			paramValue, _ := parsePrimitive(rawParamValue, field.Type)
			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	query = newQueryStruct.Interface().(T_Query)

	return query
}

func (q RouteQuery[T_Query]) Validate() (T_Query, error) {
	var query T_Query
	queryType := reflect.TypeOf(query)
	if queryType == nil {
		return query, nil
	}

	newQueryStruct := reflect.New(queryType).Elem()

	var error error

	for i := 0; i < queryType.NumField(); i++ {
		field := queryType.Field(i)
		structField := newQueryStruct.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := q.Request.URL.Query().Get(field.Name)
			paramValue, err := parsePrimitive(rawParamValue, field.Type)
			if err != nil {
				error = err
				continue
			}

			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	query = newQueryStruct.Interface().(T_Query)

	return query, error
}

type RoutePath[T_Path any] struct {
	Request *http.Request
}

func (p RoutePath[T_Path]) Path() T_Path {
	var path T_Path
	pathType := reflect.TypeOf(path)
	if pathType == nil {
		return path
	}

	newPathStruct := reflect.New(pathType).Elem()

	for i := 0; i < pathType.NumField(); i++ {
		field := pathType.Field(i)
		structField := newPathStruct.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := p.Request.PathValue(field.Name)
			paramValue, _ := parsePrimitive(rawParamValue, field.Type)
			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	path = newPathStruct.Interface().(T_Path)

	return path
}

func (p RoutePath[T_Path]) Validate() (T_Path, error) {
	var path T_Path
	pathType := reflect.TypeOf(path)
	if pathType == nil {
		return path, nil
	}

	newPathStruct := reflect.New(pathType).Elem()

	var error error

	for i := 0; i < pathType.NumField(); i++ {
		field := pathType.Field(i)
		structField := newPathStruct.FieldByName(field.Name)
		if structField.CanSet() {
			rawParamValue := p.Request.PathValue(field.Name)
			paramValue, err := parsePrimitive(rawParamValue, field.Type)
			if err != nil {
				error = err
				continue
			}

			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	path = newPathStruct.Interface().(T_Path)

	return path, error
}

type RouteBody[T_Body any] struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (b RouteBody[T_Body]) Body() T_Body {
	var body T_Body
	defer b.Request.Body.Close()
	json.NewDecoder(b.Request.Body).Decode(&body)
	return body
}

func (b RouteBody[T_Body]) Validate() (T_Body, error) {
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

type RouteResponse[T_Response any] struct {
	ResponseWriter http.ResponseWriter
}

func (r RouteResponse[T_Response]) Response(data T_Response, status ...int) error {
	r.ResponseWriter.Header().Add("Content-Type", "application/json")

	if len(status) > 0 {
		r.ResponseWriter.WriteHeader(status[0])
	}

	if err := json.NewEncoder(r.ResponseWriter).Encode(data); err != nil {
		r.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}
