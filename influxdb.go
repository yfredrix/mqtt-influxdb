package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func creatTLSConfigInflux() *tls.Config {
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = false
	// Load the CA certificate
	caCert, err := x509.SystemCertPool()
	if err != nil {
		fmt.Printf("Failed to load system cert pool: %v\n", err)
		panic(err)
	}
	tlsConfig.RootCAs = caCert
	return tlsConfig
}

func influxClient(cfg config) influxdb2.Client {
	var clientOptions = influxdb2.DefaultOptions()

	clientOptions.SetApplicationName("p1DataWriterGo")
	clientOptions.SetTLSConfig(creatTLSConfigInflux())

	client := influxdb2.NewClientWithOptions(cfg.influxURL, cfg.influxToken, clientOptions)
	return client
}

func writePoint(topic string, payload InfluxMessage, client influxdb2.Client, organization string) {
	writeAPI := client.WriteAPI(organization, topic)
	p := influxdb2.NewPoint(payload.Measurement, payload.Tags, payload.Fields, payload.Time)
	writeAPI.WritePoint(p)
	writeAPI.Flush()
}
