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
	flag.Parse()

	if *apiKey == "" || *pwsName == "" {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments. APIKey and PWSName are mandatory arguments. Usage:\n")
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

	if *fieldList != "" {
		buildMap(*fieldList, &res.WeatherMessage, *jsonTags)
	}

	if *debug {
		spew.Dump(res.WeatherMessage)
	}
}
