// influxdb/influxdb.go

package influxdb

import (
	"log"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

var influxDBClient client.Client

const (
	InfluxURL  = "http://influxdb:8086"
	InfluxDB   = "metrics"
	InfluxUser = "grafana"
	InfluxPass = "grafana"
)

// Init initializes the InfluxDB client
func Init() {
	var err error
	influxDBClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     InfluxURL,
		Username: InfluxUser,
		Password: InfluxPass,
	})
	if err != nil {
		log.Fatal("Error initializing InfluxDB client:", err)
	}
}

// Close closes the InfluxDB client
func Close() {
	influxDBClient.Close()
}

// WriteToInfluxDB writes metrics to InfluxDB
func WriteToInfluxDB(measurement string, tags map[string]string, fields map[string]interface{}) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  InfluxDB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	pt, err := client.NewPoint(measurement, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	bp.AddPoint(pt)
	if err := influxDBClient.Write(bp); err != nil {
		log.Fatal(err)
	}
}
