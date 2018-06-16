package main

import (
	"log"

	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type WeatherResult struct {
	WeatherMessage CurrentConditions
	Error          error
}

var apiKey = flag.String("APIKey", "", "WeatherUnderground API key")
var pwsName = flag.String("PWSName", "", "PWS Name")
var fieldList = flag.String("FieldList", "", "List of WU attributes")
var debug = flag.Bool("Debug", false, "Dump all WU API responses")
var jsonTags = flag.Bool("JsonTags", true, "Use JSON Tags for InfluxDB fields")
var influxDBHost = flag.String("InfluxDBHost", "http://localhost:8086", "InfluxDB host name")
var influxDBName = flag.String("InfluxDBName", "weather", "InfluxDB database name")
var influxDBUser = flag.String("InfluxDBUser", "", "InfluxDB username")
var influxDBPassword = flag.String("InfluxDBPassword", "", "InfluxDB password")

func main() {
	flag.Parse()

	if *apiKey == "" || *pwsName == "" || *fieldList == "" {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments. APIKey, PWSName, fieldList are mandatory arguments. Usage:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	wuClient, err := NewClient(*pwsName, *apiKey)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan *WeatherResult)
	go func() {
		defer close(ch)
		cond, err := wuClient.GetConditions()
		ch <- &WeatherResult{cond, err}
	}()

	res := <-ch
	if res.Error != nil {
		log.Fatal(res.Error)
	}

	if *debug {
		fmt.Fprintf(os.Stderr, "Dumping WU API response structure:\n%v\n", spew.Sdump(res.WeatherMessage))
	}

	fields := buildMap(&res.WeatherMessage)

	if *debug {
		fmt.Fprintf(os.Stderr, "Dumping InfluxDB fields structure:\n%q\n\nWill not publish to InfluxDB in debug mode. Exiting.\n",
			fields)
		os.Exit(1)
	}

	c := InfluxDBClient()
	defer c.Close()

	InfluxDBPublishPoints(c, fields)
}
