package fox

import (
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 检测是否存在该 key 对应的字段
func ExistsField(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (newKey string, newValue interface{}, err error) {
	isExists := refValModel.FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == strings.ToLower(key)
	})

	if isExists.IsValid() {
		return key, value, nil
	}
	return "", "", nil
}

// 将 key 转化为小写
func KeyToLower(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (newKey string, newValue interface{}, err error) {
	return strings.ToLower(key), value, nil
}

// 将 key 转化为大写
func KeyToUpper(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (newKey string, newValue interface{}, err error) {
	return strings.ToUpper(key), value, nil
}

// 将 value 转化为对应的类型
func StringConvert(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (string, interface{}, error) {
	var err error
	var val interface{}

	refValTargetType := refValModel.FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == strings.ToLower(key)
	})

	if !refValTargetType.IsValid() {
		return key, value, nil
	}

	sValue := value[0]
	switch refValTargetType.Kind() {
	case reflect.Bool:
		if val, err = strconv.ParseBool(sValue); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Int:
		if val, err = strconv.ParseInt(sValue, 10, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Int8:
		if val, err = strconv.ParseInt(sValue, 10, 8); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Int16:
		if val, err = strconv.ParseInt(sValue, 10, 16); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Int32:
		if val, err = strconv.ParseInt(sValue, 10, 32); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Int64:
		if val, err = time.ParseDuration(sValue); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		} else if val, err = strconv.ParseInt(sValue, 10, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uint:
		if val, err = strconv.ParseUint(sValue, 10, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uint8:
		if val, err = strconv.ParseUint(sValue, 10, 8); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uint16:
		if val, err = strconv.ParseUint(sValue, 10, 16); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uint32:
		if val, err = strconv.ParseUint(sValue, 10, 32); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uint64:
		if val, err = strconv.ParseUint(sValue, 10, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Uintptr:
		if val, err := strconv.ParseUint(sValue, 10, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Float32:
		if val, err = strconv.ParseFloat(sValue, 32); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Float64:
		if val, err = strconv.ParseFloat(sValue, 64); err == nil {
			return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.String:
		return key, sValue, nil
	case reflect.Slice:
		ref := reflect.New(refValTargetType.Type())
		ref.Elem().Set(reflect.MakeSlice(refValTargetType.Type(), 0, 0))
		switch refValTargetType.Type().Elem().Kind() {
		case reflect.Uint8:
			if val, err = hex.DecodeString(sValue); err == nil {
				return key, reflect.ValueOf(val).Convert(refValTargetType.Type()).Interface(), nil
			}
		default:
			if sValue != "" && sValue != "[]" {
				if err = json.Unmarshal([]byte(sValue), ref.Interface()); err != nil {
					return "", nil, nil
				}
			}
			return key, ref.Elem().Convert(refValTargetType.Type()).Interface(), nil
		}
	case reflect.Map:
		ref := reflect.New(refValTargetType.Type())
		ref.Elem().Set(reflect.MakeMap(refValTargetType.Type()))
		if sValue != "" && sValue != "{}" {
			if err = json.Unmarshal([]byte(sValue), ref.Interface()); err != nil {
				return "", nil, nil
			}
		}
		return key, ref.Elem().Convert(refValTargetType.Type()).Interface(), nil
	case reflect.Struct:
		ref := reflect.New(refValTargetType.Type())
		if sValue != "" && sValue != "{}" {
			if err = json.Unmarshal([]byte(sValue), ref.Interface()); err != nil {
				return "", nil, nil
			}
		}
		return key, ref.Elem().Interface(), nil
	case reflect.Ptr:
		return key, reflect.New(refValTargetType.Type().Elem()), nil
	}
	if err != nil {
		return "", nil, ErrorType
	}
	return "", nil, ErrorUnsupportedType
}
