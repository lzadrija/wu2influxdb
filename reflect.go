package main

import (
	"fmt"
	"reflect"
	"strings"
)

func buildMap(r *CurrentConditions) map[string]interface{} {
	m := make(map[string]interface{})
	v := reflect.ValueOf(m)
	resWeather := r.CurrentWeather
	vW := reflect.ValueOf(resWeather)
	tW := reflect.TypeOf(resWeather)

	for _, kv := range strings.Split(*fieldList, ",") {
		// Try to directly lookup by field name
		f := vW.FieldByName(kv)
		if t, ok := tW.FieldByName(kv); ok && f.IsValid() {
			if *jsonTags {
				v.SetMapIndex(reflect.ValueOf(t.Tag.Get("json")), f)
			} else {
				v.SetMapIndex(reflect.ValueOf(kv), f)
			}
		}

		// Try to lookup through JSON tags
		reflectRecursive(&resWeather, &kv, v)
	}

	// Always pass Unix timestamp
	m["observation_epoch"] = resWeather.ObservationEpoch
	return m
}

func reflectRecursive(s interface{}, tagName *string, v reflect.Value) {
	rType := reflect.TypeOf(s).Elem()
	rValue := reflect.ValueOf(s).Elem()

	for i := 0; i < rType.NumField(); i++ {
		tName := rType.Field(i).Name
		tTag := rType.Field(i).Tag
		vValue := rValue.Field(i).Interface()
		vAddr := rValue.Field(i).Addr()

		switch rValue.Field(i).Kind() {
		case reflect.Struct:
			reflectRecursive(vAddr.Interface(), tagName, v)
		case reflect.Ptr:
			if vValue != nil {
				reflectRecursive(vValue, tagName, v)
			}
		default:
			if tag := tTag.Get("json"); tag == *tagName {
				s := fmt.Sprintf("%v", vValue)
				if *jsonTags {
					v.SetMapIndex(reflect.ValueOf(tTag.Get("json")), reflect.ValueOf(s))
				} else {
					v.SetMapIndex(reflect.ValueOf(tName), reflect.ValueOf(s))
				}
			}
		}
	}
}
