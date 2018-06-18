# wu2influxdb [![GitHub license](https://img.shields.io/github/license/dkorunic/wu2influxdb.svg)](https://github.com/dkorunic/wu2influxdb/blob/master/LICENSE.txt) [![GitHub release](https://img.shields.io/github/release/dkorunic/wu2influxdb.svg)](https://github.com/dkorunic/wu2influxdb/releases/latest) [![Go Report Card](https://goreportcard.com/badge/github.com/dkorunic/wu2influxdb)](https://goreportcard.com/badge/github.com/dkorunic/wu2influxdb)


## About

Wu2influxdb is a small open source software to periodically pull PWS (Personal WeatherStation) data from Weather Underground and save into InfluxDB time series database.

Typically it is possible to export any data that [Weather Underground API](https://www.wunderground.com/weather/api/d/docs) provides for any PWS and save into InfluxDB database, optionally charting it in [Grafana](https://grafana.com/) later on.

JSON attributes that can be polled and saved are (not all make sense to save in InfluxDB however):

* UV
* city
* city
* conditions
* country
* country
* country\_iso3166
* country\_iso3166
* current\_observation
* dewpoint\_c
* dewpoint\_f
* dewpoint\_string
* display\_location
* elevation
* elevation
* estimated
* features
* feelslike\_c
* feelslike\_f
* feelslike\_string
* forecast\_url
* full
* full
* heat\_index\_c
* heat\_index\_f
* heat\_index\_string
* history\_url
* icon
* icon\_url
* image
* latitude
* latitude
* link
* local\_epoch
* local\_time\_rfc822
* local\_tz\_long
* local\_tz\_offset
* local\_tz\_short
* longitude
* longitude
* magic
* nowcast
* ob\_url
* observation\_epoch
* observation\_location
* observation\_time
* observation\_time\_rfc822
* precip\_1hr\_in
* precip\_1hr\_metric
* precip\_1hr\_string
* precip\_today\_in
* precip\_today\_metric
* precip\_today\_string
* pressure\_in
* pressure\_mb
* pressure\_trend
* relative\_humidity
* response
* solarradiation
* state
* state
* state\_name
* station\_id
* temp\_c
* temp\_f
* temperature\_string
* termsofService
* title
* url
* version
* visibility\_km
* visibility\_mi
* weather
* wind\_degrees
* wind\_dir
* wind\_gust\_kph
* wind\_gust\_mph
* wind\_kph
* wind\_mph
* wind\_string
* windchill\_c
* windchill\_f
* windchill\_string
* wmo
* zip

It is possible to use both native Weather Underground JSON attribute naming and naming from the Golang code. It is also possible to export data with WU JSON attributes field naming or with pretty Golang naming.

## Requirements

* InfluxDB database access with r/w privileges (database name, username, password): use **InfluxDBHost**, **InfluxDBName** and **InfluxDBPassword** parameters
* Weather Underground API key: **APIKey** parameter
* Weather Underground PWS name: **PWSName** parameter
* list of WU attributes to import: **FieldList** parameter

## Installation

There are two ways of installing wu2influxdb:

### Manual

Download your preferred flavor from [the releases](https://github.com/dkorunic/wu2influxdb/releases/latest) page and install manually.

### Using go get

```shell
go get https://github.com/dkorunic/wu2influxdb
```

## Usage

```shell
Usage of wu2influxdb:
  -APIKey string
    	WeatherUnderground API key
  -Debug
    	Dump all WU API responses
  -FieldList string
    	List of WU attributes
  -InfluxDBHost string
    	InfluxDB host name (default "http://localhost:8086")
  -InfluxDBName string
    	InfluxDB database name
  -InfluxDBPassword string
    	InfluxDB password
  -InfluxDBUser string
    	InfluxDB username
  -JsonTags
    	Use WU JSON names for InfluxDB fields (default true)
  -PWSName string
    	PWS Name
```

Typical use example to poll IZAGREB51 PWS data and import some of the data into InfluxDB weather database is:

```shell
wu2influxdb -APIKey XXXXXXXXXXXXXXXX \
  -PWSName IZAGREB51 \
  -FieldList temp_c,dewpoint_c,relative_humidity,pressure_mb,wind_kph,solarradiation,precip_today_metric,precip_1hr_metric \
  -InfluxDBName weather
```

To debug remote responses, we can use Debug parameter:

```shell
wu2influxdb -Debug \
  -APIKey XXXXXXXXXXXXXXXX \
  -PWSName IZAGREB51 \
  -FieldList temp_c,dewpoint_c,relative_humidity,pressure_mb,wind_kph,solarradiation,precip_today_metric,precip_1hr_metric
```
