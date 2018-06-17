package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	WuAPIPWSURL         = "http://api.wunderground.com/api/%s/conditions/q/pws:%s.json"
	DefaultWuAPITimeout = time.Second * 15
)

// WU API Conditions Response
type CurrentConditions struct {
	CurrentWeather WeatherResponse `json:"current_observation"`
	Response       Response        `json:"response"`
}

// WU API Status Response
type Response struct {
	TermsOfService string        `json:"termsofService"`
	Version        string        `json:"version"`
	ErrorResponse  ErrorResponse `json:"error"`
}

// WU API Error Response
type ErrorResponse struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// WU API Observation Response
type WeatherResponse struct {
	ObservationLocation ObservationLocation `json:"observation_location"`
	DisplayLocation     DisplayLocation     `json:"display_location"`
	Description         string              `json:"weather"`
	*Temperature
	*Precipitation
	*Wind
	*Windchill
	*Dewpoint
	*Pressure
	*Solar
	*Visibility
	*ObservationTimeStamp
}

type ObservationLocation struct {
	City           string      `json:"city"`
	Full           string      `json:"full"`
	Elevation      string      `json:"elevation"`
	Country        string      `json:"country"`
	Longitude      json.Number `json:"longitude"`
	State          string      `json:"state"`
	CountryISO3166 string      `json:"country_iso3166"`
	Latitude       json.Number `json:"latitude"`
}

type DisplayLocation struct {
	City           string      `json:"city"`
	Full           string      `json:"full"`
	Magic          json.Number `json:"magic"`
	StateName      string      `json:"state_name"`
	Zip            json.Number `json:"zip"`
	Country        string      `json:"country"`
	Longitude      json.Number `json:"longitude"`
	State          json.Number `json:"state"`
	Wmo            json.Number `json:"wmo"`
	CountryISO3166 string      `json:"country_iso3166"`
	Latitude       json.Number `json:"latitude"`
	Elevation      json.Number `json:"elevation"`
}

type Temperature struct {
	Description         string      `json:"temperature_string"`
	HeatIndexString     string      `json:"heat_index_string"`
	Fahrenheit          json.Number `json:"temp_f"`
	Celsius             json.Number `json:"temp_c"`
	FeelsLikeFahrenheit json.Number `json:"feelslike_f"`
	HeatIndexFahrenheit string      `json:"heat_index_f"`
	FeelsLikeCelsius    json.Number `json:"feelslike_c"`
	HeatIndexCelsius    string      `json:"heat_index_c"`
}

type Precipitation struct {
	Description       string      `json:"precip_today_string"`
	PrecipTodayMetric json.Number `json:"precip_today_metric"`
	PrecipTodayIn     json.Number `json:"precip_today_in"`
	Precip1HrString   string      `json:"precip_1hr_string"`
	Precip1HrMetric   json.Number `json:"precip_1hr_metric"`
	Precip1HrIn       json.Number `json:"precip_1hr_in"`
	RelativeHumidity  string      `json:"relative_humidity"`
}

type Wind struct {
	Description string      `json:"wind_string"`
	Direction   string      `json:"wind_dir"`
	Degrees     json.Number `json:"wind_degrees"`
	MPH         json.Number `json:"wind_mph"`
	GustMPH     json.Number `json:"wind_gust_mph"`
	KPH         json.Number `json:"wind_kph"`
	GustKPH     json.Number `json:"wind_gust_kph"`
}

type Windchill struct {
	Description string `json:"windchill_string"`
	Fahrenheit  string `json:"windchill_f"`
	Celsius     string `json:"windchill_c"`
}

type Dewpoint struct {
	Description string      `json:"dewpoint_string"`
	Fahrenheit  json.Number `json:"dewpoint_f"`
	Celsius     json.Number `json:"dewpoint_c"`
}

type Pressure struct {
	Trend string      `json:"pressure_trend"`
	IN    json.Number `json:"pressure_in"`
	MB    json.Number `json:"pressure_mb"`
}

type Solar struct {
	Radiation json.Number `json:"solarradiation"`
	UV        json.Number `json:"UV"`
}

type Visibility struct {
	VisibilityKM json.Number `json:"visibility_km"`
	VisibilityMI json.Number `json:"visibility_mi"`
}

type ObservationTimeStamp struct {
	ObservationTime       string      `json:"observation_time"`
	ObservationEpoch      json.Number `json:"observation_epoch"`
	ObservationTimeRFC822 string      `json:"observation_time_rfc822"`
}

type Client struct {
	httpClient *http.Client
	wuURL      *url.URL
}

// Prepare HTTP client structure for WU API request
func NewClient(PWSName string, APIKey string) (*Client, error) {
	wuURL, err := url.Parse(fmt.Sprintf(WuAPIPWSURL, APIKey, PWSName))
	if err != nil {
		log.Fatal(err)
	}

	c := &Client{httpClient: &http.Client{Timeout: DefaultWuAPITimeout}, wuURL: wuURL}
	return c, nil
}

// Get current conditions from WU API for PWS
func (c *Client) GetConditions() (CurrentConditions, error) {
	req, err := http.NewRequest("GET", c.wuURL.String(), nil)
	if err != nil {
		return CurrentConditions{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CurrentConditions{}, err
	}
	defer resp.Body.Close()

	// Fetch whole body at once
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CurrentConditions{}, err
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = fmt.Errorf(string(body))
		return CurrentConditions{}, err
	}

	// Dump HTTP body response if debugging
	if *debug {
		fmt.Fprintf(os.Stderr, "Dumping raw WU API:\n%v\n", string(body))
	}

	// Parse JSON
	cond := CurrentConditions{}
	err = ffjson.Unmarshal(body, &cond)
	if err != nil {
		return cond, err
	}

	// Handle WU API Errors
	errRes := cond.Response.ErrorResponse
	if errRes.Type != "" || errRes.Description != "" {
		return cond, fmt.Errorf("error from WU API: Type \"%s\", Description \"%s\"",
			errRes.Type, errRes.Description)
	}

	return cond, nil
}
