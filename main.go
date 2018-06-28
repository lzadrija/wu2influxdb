package main

import (
	"log"
	"net/url"
	"regexp"
	"strings"

	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type weatherResult struct {
	WeatherMessage CurrentConditions
	Error          error
}

const (
	defaultAPIKeyRegex  = `^[a-zA-Z0-9]{16}$`
	defaultPwsNameRegex = `^[a-zA-Z0-9_-]+$`
)

func main() {
	apiKey := flag.String("APIKey", "", "WeatherUnderground API key")
	pwsName := flag.String("PWSName", "", "PWS Name")
	fieldList := flag.String("FieldList", "", "List of WU attributes")
	debug := flag.Bool("Debug", false, "Dump all WU API responses")
	jsonTags := flag.Bool("JsonTags", true, "Use WU JSON names for InfluxDB fields")
	influxDBHost := flag.String("InfluxDBHost", "http://localhost:8086", "InfluxDB host name")
	influxDBName := flag.String("InfluxDBName", "", "InfluxDB database name")
	influxDBUser := flag.String("InfluxDBUser", "", "InfluxDB username")
	influxDBPassword := flag.String("InfluxDBPassword", "", "InfluxDB password")
	flag.Parse()
	flagsCheck(debug, apiKey, pwsName, fieldList, influxDBHost, influxDBName)

	wuClient, err := NewWuClient(pwsName, apiKey, debug)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan *weatherResult)
	go func() {
		defer close(ch)
		cond, err := wuClient.GetWuConditions()
		ch <- &weatherResult{cond, err}
	}()

	res := <-ch
	if res.Error != nil {
		log.Fatal(res.Error)
	}

	if *debug {
		fmt.Fprintf(os.Stderr, "Dumping WU API response structure:\n%v\n", spew.Sdump(res.WeatherMessage))
	}

	// prepare InfluxDB metric fields
	fieldValuesByName := buildMetricsFields(fieldList, jsonTags, &res.WeatherMessage)

	if *debug {
		fmt.Fprintf(os.Stderr,
			"Dumping InfluxDB fields structure:\n%v\n\nWill not publish to InfluxDB in debug mode. Exiting.\n", fieldValuesToFloat(fieldValuesByName))
		os.Exit(1)
	}

	influxDbClient := InfluxDBClient(influxDBHost, influxDBUser, influxDBPassword)
	defer influxDbClient.Close()

	InfluxDBPublishPoints(influxDbClient, fieldValuesByName, influxDBName, pwsName)
}

func flagsCheck(debug *bool, apiKey, pwsName, fieldList, influxDBHost, influxDBName *string) {
	if *apiKey == "" || *pwsName == "" || *fieldList == "" {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments. APIKey, PWSName and fieldList are mandatory parameters.\n")
		flag.Usage()
		os.Exit(1)
	}

	if regexp.MustCompile(defaultAPIKeyRegex).MatchString(*apiKey) == false {
		fmt.Fprintf(os.Stderr, "APIKey parameter \"%s\" is not in valid format (16-digit alphanumeric string required).\n", *apiKey)
		os.Exit(1)
	}

	if regexp.MustCompile(defaultPwsNameRegex).MatchString(*pwsName) == false {
		fmt.Fprintf(os.Stderr, "PWSName parameter \"%s\" is not in valid format (alphanumeric string including minus and underscore required).\n", *pwsName)
		os.Exit(1)
	}

	if *influxDBHost != "" {
		p, err := url.Parse(*influxDBHost)
		if err != nil {
			log.Fatal(err)
		}

		if p.Host == "" {
			fmt.Fprintf(os.Stderr, "InfluxDBHost parameter \"%s\" is missing proper host name.\n", *influxDBHost)
			os.Exit(1)
		}

		// URL path has been specified and InfluxDBName argument is not there, use it as InfluxDB database name
		if p.Path != "" {
			db := strings.TrimSuffix(strings.TrimPrefix(p.Path, "/"), "/")

			if db != "" && *influxDBName == "" {
				*influxDBName = db
			}
		}
	}

	// We are not in debug mode and InfluxDBName is empty
	if !*debug && *influxDBName == "" {
		fmt.Fprintf(os.Stderr, "InfluxDBName parameter is missing and it is a mandatory parameter when not in debug mode.\n")
		flag.Usage()
		os.Exit(1)
	}
}
