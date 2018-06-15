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

func main() {
	apiKey := flag.String("APIKey", "", "WeatherUnderground API key")
	pwsName := flag.String("PWSName", "", "PWS Name")
	fieldList := flag.String("FieldList", "", "List of WU attributes")
	debug := flag.Bool("Debug", false, "Dump all WU API responses")
	jsonTags := flag.Bool("JsonTags", true, "Use JSON Tags for InfluxDB fields")
	influxDBHost := flag.String("InfluxDBHost", "http://localhost:8086", "InfluxDB host name")
	influxDBName := flag.String("InfluxDBName", "weather", "InfluxDB database name")
	influxDBUser := flag.String("InfluxDBUser", "", "InfluxDB username")
	influxDBPassword := flag.String("InfluxDBPassword", "", "InfluxDB password")

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
		fmt.Printf("Dumping WU API response structure")
		spew.Dump(res.WeatherMessage)
	}

	f := buildMap(*fieldList, &res.WeatherMessage, *jsonTags)

	if *debug {
		fmt.Printf("Dumping InfluxDB fields structure: %q\n", f)
	}

	c := InfluxDBClient(influxDBHost, influxDBUser, influxDBPassword)
	defer c.Close()

	InfluxDBPublishPoints(c, influxDBName, f, pwsName)
}
