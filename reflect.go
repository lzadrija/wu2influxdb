package main

import (
	"fmt"
	"reflect"
	"strings"
)

// buildMetricsFields prepares a list of fields for InfluxDB fields metrics by lookuping both native GO field names or
// JSON field names
func buildMetricsFields(fieldList *string, jsonTags *bool, currentConditions *CurrentConditions) map[string]interface{} {
	fieldValuesByName := make(map[string]interface{})
	valueOfFieldValuesByName := reflect.ValueOf(fieldValuesByName)

	weatherResponse := currentConditions.CurrentWeather
	weatherResponseValue := reflect.ValueOf(weatherResponse)
	weatherResponseType := reflect.TypeOf(weatherResponse)

	for _, fieldName := range strings.Split(*fieldList, ",") {
		// Try to directly lookup by field name
		fieldValue := weatherResponseValue.FieldByName(fieldName)
		if fieldDescription, isFieldFound := weatherResponseType.FieldByName(fieldName); isFieldFound && fieldValue.IsValid() {

			valueOfFieldValuesByName.SetMapIndex(reflect.ValueOf(getFieldName(jsonTags, fieldDescription.Tag.Get("json"), fieldName)),
				reflect.ValueOf(fieldValue))
		}

		// Try to lookup through JSON tags
		buildMetricsFieldsByRecursiveFieldNameSearch(jsonTags, &weatherResponse, &fieldName, valueOfFieldValuesByName)
	}

	// Always pass Unix timestamp
	fieldValuesByName["observation_epoch"] = weatherResponse.ObservationEpoch

	return fieldValuesByName
}

// buildMetricsFieldsByRecursiveFieldNameSearch recursively searches for a list of JSON field names and builds a list of fields for InfluxDB fields,
// either using native GO field names or JSON field names
func buildMetricsFieldsByRecursiveFieldNameSearch(jsonTags *bool, weatherResponse interface{}, fieldNameToFind *string, valueOfFieldValuesByName reflect.Value) {
	weatherResponseType := reflect.TypeOf(weatherResponse).Elem()
	weatherResponseValue := reflect.ValueOf(weatherResponse).Elem()

	for i := 0; i < weatherResponseType.NumField(); i++ {
		weatherResponseFieldValue := weatherResponseValue.Field(i)

		switch weatherResponseFieldValue.Kind() {
		case reflect.Struct:
			buildMetricsFieldsByRecursiveFieldNameSearch(jsonTags, weatherResponseFieldValue.Addr().Interface(), fieldNameToFind, valueOfFieldValuesByName)
		case reflect.Ptr:
			if weatherResponseFieldValue.Interface() != nil {
				buildMetricsFieldsByRecursiveFieldNameSearch(jsonTags, weatherResponseFieldValue.Interface(), fieldNameToFind, valueOfFieldValuesByName)
			}
		default:
			fieldTagName := weatherResponseType.Field(i).Tag.Get("json")

			if fieldTagName == *fieldNameToFind {
				fieldValue := fmt.Sprintf("%v", weatherResponseFieldValue.Interface())

				valueOfFieldValuesByName.SetMapIndex(reflect.ValueOf(getFieldName(jsonTags, fieldTagName, weatherResponseType.Field(i).Name)),
					reflect.ValueOf(fieldValue))
			}
		}
	}
}

func getFieldName(jsonTags *bool, fieldTagName string, fieldName string) string {

	if *jsonTags {
		return fieldTagName
	}

	return fieldName
}