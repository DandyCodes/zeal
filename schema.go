package zeal

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/a-h/rest"
)

func registerRoute(mux *ZealMux, pattern string, routeType reflect.Value) {
	route, err := newRoute(pattern, mux)
	if err != nil {
		fmt.Println(err)
		return
	}

	if routeType.Kind() == reflect.Interface {
		registerResponse(route, nil)
		return
	}

	paramsTypeName := getTypeName(HasParams[any]{})
	paramsField := routeType.FieldByName(paramsTypeName)
	if paramsField.IsValid() {
		method := paramsField.Addr().MethodByName("Params")
		if err := registerParams(route, pattern, method.Type().Out(0)); err != nil {
			fmt.Println(err)
		}
	}

	bodyTypeName := getTypeName(HasBody[any]{})
	bodyField := routeType.FieldByName(bodyTypeName)
	if bodyField.IsValid() {
		method := bodyField.Addr().MethodByName("Body")
		registerBody(route, method.Type().Out(0))
	}

	responseTypeName := getTypeName(HasResponse[any]{})
	responseField := routeType.FieldByName(responseTypeName)
	if responseField.IsValid() {
		method := responseField.Addr().MethodByName("Response")
		registerResponse(route, method.Type().In(1))
	} else {
		registerResponse(route, nil)
	}
}

func newRoute(pattern string, mux *ZealMux) (*rest.Route, error) {
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

func registerParams(route *rest.Route, pattern string, paramsType reflect.Type) error {
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

type urlPathParamPlaceholder struct {
	name             string
	validationRegexp string
}

func getURLPathParamPlaceholder(urlSlug string) (urlPathParamPlaceholder, error) {
	placeholder := urlPathParamPlaceholder{}
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

func registerBody(route *rest.Route, bodyType reflect.Type) {
	if bodyType == nil {
		return
	}

	route.HasRequestModel(rest.Model{Type: bodyType})
}

func registerResponse(route *rest.Route, responseType reflect.Type) {
	if responseType == nil {
		route.HasResponseModel(http.StatusOK, rest.Model{Type: reflect.TypeOf("")})
		return
	}

	route.HasResponseModel(http.StatusOK, rest.Model{Type: responseType})
}
