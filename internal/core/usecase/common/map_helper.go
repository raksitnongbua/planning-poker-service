package common

import "reflect"

func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)

	for i := 0; i < objValue.NumField(); i++ {
		fieldValue := objValue.Field(i)
		fieldTag := objType.Field(i).Tag.Get("json")
		result[fieldTag] = fieldValue.Interface()
	}

	return result
}
