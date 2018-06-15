package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultInfluxDBPrecision   = "s"
	DefaultInfluxDBMeasurement = "climate"
)

// Initiate connection to InfluxDB service
func InfluxDBClient(influxDBHost, influxDBUser, influxDBPassword *string) client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: *influxDBHost, Username: *influxDBUser,
		Password: *influxDBPassword})
	if err != nil {
		log.Fatal(err)
	}

	return c
}

// Publish batched points to InfluxDB database
func InfluxDBPublishPoints(c client.Client, influxDBName *string, fields map[string]interface{}, pwsName *string) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: *influxDBName, Precision: DefaultInfluxDBPrecision})
	if err != nil {
		log.Fatal(err)
	}

	tags := map[string]string{
		"source":   "wunderground",
		"pws_name": *pwsName,
	}

	// convert Unix epoch into valid InfluxDB timestamp
	unixTime, ok := fields["observation_epoch"]
	if !ok {
		log.Fatal(fmt.Errorf("missing observation_epoch timestamp in fields structure"))
	}
	i, err := strconv.ParseInt(unixTime.(string), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	pointTime := time.Unix(i, 0)

	// Convert all values to float
	floatifyFields(fields)

	pt, err := client.NewPoint(DefaultInfluxDBMeasurement, tags, fields, pointTime)
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}

	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}

// Converts all string values in map to float64
func floatifyFields(fields map[string]interface{}) {
	for k, j := range fields {
		var f string

		// Convert to strings
		switch v := j.(type) {
		case string:
			f = j.(string)
		case fmt.Stringer:
			f = v.String()
		default:
			f = fmt.Sprintf("%v", v)
		}

		// Trim whitespace and percentage
		f = strings.TrimSpace(strings.TrimSuffix(f, "%"))

		// Convert to float64 if possible
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			continue
		}
		fields[k] = v

	}
}
