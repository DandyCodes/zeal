package zeal

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/a-h/rest"
)

type RouteSchema struct {
	pattern      string
	paramsType   reflect.Type
	bodyType     reflect.Type
	responseType reflect.Type
}

func getRouteSchema[ResponseType, ParamsType, BodyType any](pattern string) RouteSchema {
	var params ParamsType
	var body BodyType
	var response ResponseType
	return RouteSchema{pattern: pattern, paramsType: reflect.TypeOf(params), bodyType: reflect.TypeOf(body), responseType: reflect.TypeOf(response)}
}

func registerRequestModel(route *rest.Route, bodyType reflect.Type) {
	if bodyType != nil {
		route.HasRequestModel(rest.Model{Type: bodyType})
	}
}

func registerResponseModel(route *rest.Route, responseType reflect.Type) {
	if responseType != nil {
		route.HasResponseModel(http.StatusOK, rest.Model{Type: responseType})
	} else {
		route.HasResponseModel(http.StatusOK, rest.Model{Type: reflect.TypeOf("")})
	}
}

func isPathParam(s string, path map[string]rest.PathParam) bool {
	for pathParamName := range path {
		if pathParamName == s {
			return true
		}
	}
	return false
}

func registerParameters(route *rest.Route, pattern string, paramsType reflect.Type) {
	params, err := getParamsFromPattern(pattern)
	if err != nil {
		log.Fatal("Failed to parse path parameters from", pattern)
	}
	if paramsType != nil {
		for i := 0; i < paramsType.NumField(); i++ {
			field := paramsType.Field(i)
			primitiveSchemaType := getPrimitiveSchemaType(field.Type.Kind())
			if primitiveSchemaType == "" {
				log.Fatal("Params type must contain only primitive types", paramsType)
			}

			if isPathParam(field.Name, params.Path) {
				route.HasPathParameter(field.Name, rest.PathParam{Type: primitiveSchemaType})
			} else {
				route.HasQueryParameter(field.Name, rest.QueryParam{Type: primitiveSchemaType, Required: true})
			}
		}
	}
}

func getParamsFromPattern(s string) (p rest.Params, err error) {
	p.Path = make(map[string]rest.PathParam)
	p.Query = make(map[string]rest.QueryParam)

	u, err := url.Parse(s)
	if err != nil {
		return
	}

	// Path.
	s = u.Path
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimPrefix(s, "/")
	segments := strings.Split(s, "/")
	for _, segment := range segments {
		name, pattern, ok := getPlaceholder(segment)
		if !ok {
			continue
		}
		p.Path[name] = rest.PathParam{
			Regexp: pattern,
		}
	}

	// Query.
	q := u.Query()
	for k := range q {
		name, _, ok := getPlaceholder(q.Get(k))
		if !ok {
			continue
		}
		p.Query[name] = rest.QueryParam{
			Description: "",
			Required:    false,
			AllowEmpty:  false,
		}
	}

	return
}

func getPlaceholder(s string) (name string, pattern string, ok bool) {
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return
	}
	parts := strings.SplitN(s[1:len(s)-1], ":", 2)
	name = parts[0]
	if len(parts) > 1 {
		pattern = parts[1]
	}
	return name, pattern, true
}

func getPrimitiveSchemaType(kind reflect.Kind) rest.PrimitiveType {
	switch kind {
	case reflect.Bool:
		return rest.PrimitiveTypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rest.PrimitiveTypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rest.PrimitiveTypeInteger
	case reflect.Float32, reflect.Float64:
		return rest.PrimitiveTypeFloat64
	case reflect.String:
		return rest.PrimitiveTypeString
	default:
		return ""
	}
}
