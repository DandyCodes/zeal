package zeal

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func parsePrimitive(rawValue any, valueType reflect.Type) (any, error) {
	value := fmt.Sprintf("%v", rawValue)
	var parsed any

	switch valueType.Kind() {
	case reflect.Float32, reflect.Float64:
		val, err := parseFloat(value, valueType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := parseInt(value, valueType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := parseUint(value, valueType)
		if err != nil {
			return nil, err
		}

		parsed = val
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean from: %v", value)
		}

		parsed = val
	case reflect.String:
		parsed = value
	default:
		return nil, fmt.Errorf("unsupported type: %v", valueType.Kind())
	}

	return parsed, nil
}

func parseFloat(value string, valueType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse float from: %v", value)
	}

	switch valueType.Kind() {
	case reflect.Float32:
		if !isFloat32InRange(val) {
			return nil, fmt.Errorf("value out of range for float32: %v", val)
		}
		parsed = float32(val)
	case reflect.Float64:
		parsed = float64(val)
	default:
		return nil, fmt.Errorf("expected float, received: %v", valueType)
	}

	return parsed, nil
}

func parseInt(value string, valueType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integer from: %v", value)
	}

	if !isIntInRange(val, valueType.Kind()) {
		return nil, fmt.Errorf("value out of range for %v: %d", valueType.Kind(), val)
	}

	switch valueType.Kind() {
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
		return nil, fmt.Errorf("expected integer, received: %v", valueType)
	}

	return parsed, nil
}

func parseUint(value string, valueType reflect.Type) (any, error) {
	var parsed any

	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unsigned integer from: %v", value)
	}

	if !isUintInRange(val, valueType.Kind()) {
		return nil, fmt.Errorf("value out of range for %v: %d", valueType.Kind(), val)
	}

	switch valueType.Kind() {
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
		return nil, fmt.Errorf("expected unsigned integer, received: %v", valueType)
	}

	return parsed, nil
}

func isFloat32InRange(value float64) bool {
	return value >= -math.MaxFloat32 && value <= math.MaxFloat32
}

func isIntInRange(value int64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Int:
		return value >= math.MinInt && value <= math.MaxInt
	case reflect.Int8:
		return value >= math.MinInt8 && value <= math.MaxInt8
	case reflect.Int16:
		return value >= math.MinInt16 && value <= math.MaxInt16
	case reflect.Int32:
		return value >= math.MinInt32 && value <= math.MaxInt32
	case reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintInRange(value uint64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint:
		return value <= math.MaxUint
	case reflect.Uint8:
		return value <= math.MaxUint8
	case reflect.Uint16:
		return value <= math.MaxUint16
	case reflect.Uint32:
		return value <= math.MaxUint32
	case reflect.Uint64:
		return true
	default:
		return false
	}
}
