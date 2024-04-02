package zeal

import (
	"encoding/json"
	"net/http"
	"reflect"
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
			rawParamValue := r.PathValue(field.Name)
			if rawParamValue == "" {
				rawParamValue = r.URL.Query().Get(field.Name)
			}
			paramValue, err := parsePrimitive(rawParamValue, field.Type)
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
