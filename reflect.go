package main

import (
	"fmt"
	"reflect"
	"strings"
)

func buildMap(s string, r *CurrentConditions) error {
	m := make(map[string]string)
	v := reflect.ValueOf(m)
	resWeather := r.CurrentWeather
	vW := reflect.ValueOf(resWeather)

	for _, kv := range strings.Split(s, ",") {
		// Try to directly lookup by field name
		f := vW.FieldByName(kv)
		if f.IsValid() {
			v.SetMapIndex(reflect.ValueOf(kv), f)
		}

		// Try to lookup through JSON tags
		reflectRecursive(&resWeather, &kv, v)
	}

	fmt.Printf("%q\n", m)
	return nil
}

func reflectRecursive(s interface{}, tagName *string, v reflect.Value) {
	rType := reflect.TypeOf(s).Elem()
	rValue := reflect.ValueOf(s).Elem()

	for i := 0; i < rType.NumField(); i++ {
		tName := rType.Field(i).Name
		vValue := rValue.Field(i).Interface()
		vAddr := rValue.Field(i).Addr()

		switch rValue.Field(i).Kind() {
		case reflect.Struct:
			reflectRecursive(vAddr.Interface(), tagName, v)
		case reflect.Ptr:
			reflectRecursive(vValue, tagName, v)
		default:
			if tag, ok := rValue.Type().Field(i).Tag.Lookup("json"); ok && tag == *tagName {
				s := fmt.Sprintf("%v", vValue)
				v.SetMapIndex(reflect.ValueOf(tName), reflect.ValueOf(s))
			}
		}
	}
}
