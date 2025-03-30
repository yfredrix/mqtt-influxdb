package main

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func influxClient(cfg config) influxdb2.Client {
	var clientOptions = influxdb2.DefaultOptions()

	clientOptions.SetApplicationName("p1DataWriterGo")

	client := influxdb2.NewClientWithOptions(cfg.influxURL, cfg.influxToken, clientOptions)
	return client
}

func writePoint(topic string, payload Message, client influxdb2.Client, organization string) {
	writeAPI := client.WriteAPI(organization, topic)
	p := influxdb2.NewPoint(payload.Measurement, payload.Tags, payload.Fields, time.Now())
	writeAPI.WritePoint(p)
	writeAPI.Flush()
}
