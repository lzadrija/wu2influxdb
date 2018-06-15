package main

import (
	"log"

	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"os"
)

type WeatherResult struct {
	WeatherMessage CurrentConditions
	Error          error
}

func main() {
	apiKey := flag.String("APIKey", "", "WeatherUnderground API key")
	pwsName := flag.String("PWSName", "", "PWS Name")
	flag.Parse()

	if *apiKey == "" || *pwsName == "" {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments. All parameters are mandatory. Usage:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	wuClient, err := NewClient(*pwsName, *apiKey)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan *WeatherResult)
	go func() {
		cond, err := wuClient.GetConditions()
		ch <- &WeatherResult{cond, err}
	}()

	res := <-ch
	if res.Error != nil {
		log.Fatal(res.Error)
	}

	spew.Dump(res.WeatherMessage)
}
