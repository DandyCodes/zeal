package zeal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
)

func getParams[ParamsType any](r *http.Request) (ParamsType, error) {
	var error error
	var params ParamsType

	if reflect.TypeOf(params) == nil {
		return params, error
	}

	paramsType := reflect.TypeOf(params)

	newParamsStruct := reflect.New(paramsType).Elem()

	for i := 0; i < paramsType.NumField(); i++ {
		field := paramsType.Field(i)
		structField := newParamsStruct.FieldByName(field.Name)
		if structField.CanSet() {
			paramValue, err := getParam(r, field.Name, field.Type)
			if err != nil {
				error = err
				continue
			}
			structField.Set(reflect.ValueOf(paramValue))
		}
	}

	params = newParamsStruct.Interface().(ParamsType)

	return params, error
}

func getParam(r *http.Request, name string, paramType reflect.Type) (any, error) {
	param := r.PathValue(name)
	if param == "" {
		param = r.URL.Query().Get(name)
	}
	return parseParam(param, paramType)
}

func parseParam(param string, paramType reflect.Type) (interface{}, error) {
	var parsed interface{}

	switch paramType.Kind() {
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse number from: %v", param)
		}
		if paramType.Kind() == reflect.Float32 {
			if !isFloat32InRange(val) {
				return nil, fmt.Errorf("value out of range for float32: %v", val)
			}
			parsed = float32(val)
		} else {
			parsed = val
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer from: %v", param)
		}
		if !isIntInRange(val, paramType.Kind()) {
			return nil, fmt.Errorf("value out of range for %v: %d", paramType.Kind(), val)
		}
		switch paramType.Kind() {
		case reflect.Int:
			parsed = int(val)
		case reflect.Int8:
			parsed = int8(val)
		case reflect.Int16:
			parsed = int16(val)
		case reflect.Int32:
			parsed = int32(val)
		case reflect.Int64:
			parsed = val
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unsigned integer from: %v", param)
		}
		if !isUintInRange(val, paramType.Kind()) {
			return nil, fmt.Errorf("value out of range for %v: %d", paramType.Kind(), val)
		}
		switch paramType.Kind() {
		case reflect.Uint:
			parsed = uint(val)
		case reflect.Uint8:
			parsed = uint8(val)
		case reflect.Uint16:
			parsed = uint16(val)
		case reflect.Uint32:
			parsed = uint32(val)
		case reflect.Uint64:
			parsed = val
		}
	case reflect.Bool:
		val, err := strconv.ParseBool(param)
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean from: %v", param)
		}
		parsed = val
	case reflect.String:
		parsed = param
	default:
		return nil, fmt.Errorf("unsupported type: %v", paramType.Kind())
	}

	return parsed, nil
}

func isFloat32InRange(val float64) bool {
	return val >= -math.MaxFloat32 && val <= math.MaxFloat32
}

func isIntInRange(val int64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Int:
		return val >= math.MinInt && val <= math.MaxInt
	case reflect.Int8:
		return val >= math.MinInt8 && val <= math.MaxInt8
	case reflect.Int16:
		return val >= math.MinInt16 && val <= math.MaxInt16
	case reflect.Int32:
		return val >= math.MinInt32 && val <= math.MaxInt32
	case reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintInRange(val uint64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint:
		return val <= math.MaxUint
	case reflect.Uint8:
		return val <= math.MaxUint8
	case reflect.Uint16:
		return val <= math.MaxUint16
	case reflect.Uint32:
		return val <= math.MaxUint32
	case reflect.Uint64:
		return true
	default:
		return false
	}
}

func getBody[BodyType any](r *http.Request) (BodyType, bool) {
	var b BodyType
	bodyType := reflect.TypeOf(b)
	if bodyType == nil {
		return b, true
	}

	var body BodyType

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return b, false
	}

	return body, true
}
