package zeal

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/a-h/rest"
)

func registerRoute[T_Response, T_Params, T_Body any](pattern string, mux *ServeMux) {
	route, err := getRoute(pattern, mux)
	if err != nil {
		fmt.Println(err)
		return
	}

	var response T_Response
	var params T_Params
	var body T_Body
	registerResponseModel(route, reflect.TypeOf(response))
	if err := registerParameters(route, pattern, reflect.TypeOf(params)); err != nil {
		fmt.Println(err)
	}
	registerRequestModel(route, reflect.TypeOf(body))
}

func getRoute(pattern string, mux *ServeMux) (*rest.Route, error) {
	method, path, found := strings.Cut(pattern, " ")
	if !found {
		return nil, fmt.Errorf("expected URL pattern with HTTP method, received: %v", pattern)
	}

	var route *rest.Route

	switch method {
	case http.MethodConnect:
		route = mux.Api.Connect(path)
	case http.MethodDelete:
		route = mux.Api.Delete(path)
	case http.MethodGet:
		route = mux.Api.Get(path)
	case http.MethodHead:
		route = mux.Api.Head(path)
	case http.MethodOptions:
		route = mux.Api.Options(path)
	case http.MethodPatch:
		route = mux.Api.Patch(path)
	case http.MethodPost:
		route = mux.Api.Post(path)
	case http.MethodPut:
		route = mux.Api.Put(path)
	case http.MethodTrace:
		route = mux.Api.Trace(path)
	default:
		return nil, fmt.Errorf("expected HTTP method, received: %v", method)
	}

	return route, nil
}

func registerResponseModel(route *rest.Route, responseType reflect.Type) {
	if responseType == nil {
		route.HasResponseModel(http.StatusOK, rest.Model{Type: reflect.TypeOf("")})
		return
	}

	route.HasResponseModel(http.StatusOK, rest.Model{Type: responseType})
}

func registerParameters(route *rest.Route, pattern string, paramsType reflect.Type) error {
	if paramsType == nil {
		return nil
	}

	pathParams, err := getPathParams(pattern)
	if err != nil {
		return err
	}

	for i := 0; i < paramsType.NumField(); i++ {
		field := paramsType.Field(i)
		primitiveSchemaType, err := getPrimitiveSchemaType(field.Type.Kind())
		if err != nil {
			return err
		}

		pathParam, isPathParam := pathParams[field.Name]
		if isPathParam {
			route.HasPathParameter(
				field.Name,
				rest.PathParam{Type: primitiveSchemaType, Regexp: pathParam.Regexp},
			)
			continue
		}

		route.HasQueryParameter(
			field.Name,
			rest.QueryParam{
				Type:       primitiveSchemaType,
				Required:   true,
				AllowEmpty: false,
			},
		)
	}

	return nil
}

func getPathParams(pattern string) (map[string]rest.PathParam, error) {
	url, err := url.Parse(pattern)
	if err != nil {
		return nil, err
	}

	pathParams := make(map[string]rest.PathParam)

	urlSlugs := getUrlSlugs(*url)
	for i := range urlSlugs {
		placeholder, err := getURLPathParamPlaceholder(urlSlugs[i])
		if err != nil {
			continue
		}

		pathParams[placeholder.name] = rest.PathParam{
			Regexp: placeholder.validationRegexp,
		}
	}

	return pathParams, nil
}

func getUrlSlugs(url url.URL) []string {
	path := url.Path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Split(path, "/")
}

type URLPathParamPlaceholder struct {
	name             string
	validationRegexp string
}

func getURLPathParamPlaceholder(urlSlug string) (URLPathParamPlaceholder, error) {
	placeholder := URLPathParamPlaceholder{}
	if !strings.HasPrefix(urlSlug, "{") || !strings.HasSuffix(urlSlug, "}") {
		return placeholder, fmt.Errorf("expected URL placeholder, received: %v", urlSlug)
	}

	parts := strings.SplitN(urlSlug[1:len(urlSlug)-1], ":", 2)
	placeholder.name = parts[0]
	if len(parts) > 1 {
		placeholder.validationRegexp = parts[1]
	}

	return placeholder, nil
}

func getPrimitiveSchemaType(kind reflect.Kind) (rest.PrimitiveType, error) {
	switch kind {
	case reflect.Bool:
		return rest.PrimitiveTypeBool, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rest.PrimitiveTypeInteger, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rest.PrimitiveTypeInteger, nil
	case reflect.Float32, reflect.Float64:
		return rest.PrimitiveTypeFloat64, nil
	case reflect.String:
		return rest.PrimitiveTypeString, nil
	default:
		return "", fmt.Errorf("expected primitive kind, received: %v", kind)
	}
}

func registerRequestModel(route *rest.Route, bodyType reflect.Type) {
	if bodyType == nil {
		return
	}

	route.HasRequestModel(rest.Model{Type: bodyType})
}
