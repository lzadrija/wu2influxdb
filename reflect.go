package main

import (
	"fmt"
	"reflect"
	"strings"
)

// buildMetricsFields prepares a list of fields for InfluxDB fields metrics by lookuping both native GO field names or
// JSON field names
func buildMetricsFields(fieldList *string, jsonTags *bool, r *CurrentConditions) map[string]interface{} {
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
		rTagSearch(jsonTags, &resWeather, &kv, v)
	}

	// Always pass Unix timestamp
	m["observation_epoch"] = resWeather.ObservationEpoch
	return m
}

// rTagSearch recursively searches for a list of JSON field names and builds a list of fields for InfluxDB fields,
// either using native GO field names or JSON field names
func rTagSearch(jsonTags *bool, s interface{}, tagName *string, v reflect.Value) {
	rType := reflect.TypeOf(s).Elem()
	rValue := reflect.ValueOf(s).Elem()

	for i := 0; i < rType.NumField(); i++ {
		tName := rType.Field(i).Name
		tTag := rType.Field(i).Tag
		vValue := rValue.Field(i).Interface()
		vAddr := rValue.Field(i).Addr()

		switch rValue.Field(i).Kind() {
		case reflect.Struct:
			rTagSearch(jsonTags, vAddr.Interface(), tagName, v)
		case reflect.Ptr:
			if vValue != nil {
				rTagSearch(jsonTags, vValue, tagName, v)
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
