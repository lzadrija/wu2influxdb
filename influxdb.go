package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	defaultInfluxDBPrecision   = "s"
	defaultInfluxDBMeasurement = "climate"
)

// InfluxDBClient initiates connection to InfluxDB service
func InfluxDBClient(influxDBHost, influxDBUser, influxDBPassword *string) client.Client {
	influxDbClient, err := client.NewHTTPClient(client.HTTPConfig{Addr: *influxDBHost, Username: *influxDBUser, Password: *influxDBPassword})
	if err != nil {
		log.Fatal(err)
	}

	return influxDbClient
}

// InfluxDBPublishPoints publishes batched points to InfluxDB database
func InfluxDBPublishPoints(influxDbClient client.Client, fieldValuesByName map[string]interface{}, influxDBName, pwsName *string) {
	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{Database: *influxDBName, Precision: defaultInfluxDBPrecision})
	if err != nil {
		log.Fatal(err)
	}

	tags := map[string]string{
		"source":   "wunderground",
		"pws_name": *pwsName,
	}

	// convert Unix epoch into valid InfluxDB timestamp
	unixTime, ok := fieldValuesByName["observation_epoch"]
	if !ok {
		log.Fatal(fmt.Errorf("missing observation_epoch timestamp in fieldValuesByName structure"))
	}
	unixTimeAsInt, err := strconv.ParseInt(unixTime.(string), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	pointTime := time.Unix(unixTimeAsInt, 0)

	// Convert all values to float
	fieldValuesByName = fieldValuesToFloat(fieldValuesByName)

	point, err := client.NewPoint(defaultInfluxDBMeasurement, tags, fieldValuesByName, pointTime)
	if err != nil {
		log.Fatal(err)
	}
	batchPoints.AddPoint(point)

	if err := influxDbClient.Write(batchPoints); err != nil {
		log.Fatal(err)
	}

	if err := influxDbClient.Close(); err != nil {
		log.Fatal(err)
	}
}

// fieldValuesToFloat converts all values in map to float64 where conversion is possible
func fieldValuesToFloat(fieldValuesByName map[string]interface{}) map[string]interface{} {
	for name, valueOfInterfaceType := range fieldValuesByName {
		var valueAsString string

		switch value := valueOfInterfaceType.(type) {
		case float32:
			fieldValuesByName[name] = float64(value)
			continue
		case float64:
			continue
		case string:
			valueAsString = valueOfInterfaceType.(string)
		case fmt.Stringer:
			valueAsString = value.String()
		default:
			valueAsString = fmt.Sprintf("%v", value)
		}

		// Trim whitespace and percentage
		valueAsString = strings.TrimSpace(strings.TrimSuffix(valueAsString, "%"))

		// Convert to float64 if possible
		if valueOfTypeFloat, err := strconv.ParseFloat(valueAsString, 64); err == nil {
			fieldValuesByName[name] = valueOfTypeFloat
		}
	}

	return fieldValuesByName
}
