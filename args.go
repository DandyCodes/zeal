package zeal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
)

func getArgs[T_Params, T_Body any](request *http.Request) (T_Params, T_Body, error) {
	params, paramErr := getParams[T_Params](request)
	if paramErr != nil {
		var body T_Body
		return params, body, paramErr
	}

	body, bodyErr := getBody[T_Body](request)
	if bodyErr != nil {
		return params, body, bodyErr
	}

	return params, body, nil
}

func getParams[T_Params any](r *http.Request) (T_Params, error) {
	var params T_Params
	paramsType := reflect.TypeOf(params)
	if paramsType == nil {
		return params, nil
	}

	newParamsStruct := reflect.New(paramsType).Elem()

	var error error

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

	params = newParamsStruct.Interface().(T_Params)

	return params, error
}

func getParam(r *http.Request, paramName string, paramType reflect.Type) (any, error) {
	rawParamValue := r.PathValue(paramName)
	if rawParamValue == "" {
		rawParamValue = r.URL.Query().Get(paramName)
	}
	return parseParam(rawParamValue, paramType)
}

func parseParam(rawParamValue string, paramType reflect.Type) (any, error) {
	var parsed any

	switch paramType.Kind() {
	case reflect.Float32, reflect.Float64:
		val, err := parseParamFloat(rawParamValue, paramType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := parseParamInt(rawParamValue, paramType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := parseParamUint(rawParamValue, paramType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Bool:
		val, err := strconv.ParseBool(rawParamValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean from: %v", rawParamValue)
		}

		parsed = val
	case reflect.String:
		parsed = rawParamValue
	default:
		return nil, fmt.Errorf("unsupported type: %v", paramType.Kind())
	}

	return parsed, nil
}

func parseParamFloat(rawParamValue string, paramType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseFloat(rawParamValue, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse float from: %v", rawParamValue)
	}

	switch paramType.Kind() {
	case reflect.Float32:
		if !isFloat32InRange(val) {
			return nil, fmt.Errorf("value out of range for float32: %v", val)
		}
		parsed = float32(val)
	case reflect.Float64:
		parsed = float64(val)
	default:
		return nil, fmt.Errorf("expected float, received: %v", paramType)
	}

	return parsed, nil
}

func parseParamInt(rawParamValue string, paramType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseInt(rawParamValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integer from: %v", rawParamValue)
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
		parsed = int64(val)
	default:
		return nil, fmt.Errorf("expected integer, received: %v", paramType)
	}

	return parsed, nil
}

func parseParamUint(rawParamValue string, paramType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseUint(rawParamValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unsigned integer from: %v", rawParamValue)
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
		parsed = uint64(val)
	default:
		return nil, fmt.Errorf("expected unsigned integer, received: %v", paramType)
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

func getBody[T_Body any](r *http.Request) (T_Body, error) {
	var body T_Body
	bodyType := reflect.TypeOf(body)
	if bodyType == nil {
		return body, nil
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, err
	}

	return body, nil
}
