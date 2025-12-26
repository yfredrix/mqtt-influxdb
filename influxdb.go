package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	influxdb3 "github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
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

func influxClient(cfg config) *influxdb3.Client {
	transport := &http.Transport{
		TLSClientConfig: creatTLSConfigInflux(),
	}
	http_client := &http.Client{
		Transport: transport,
	}

	client, err := influxdb3.New(influxdb3.ClientConfig{Host: cfg.influxURL, Token: cfg.influxToken, Database: cfg.influxDatabase, HTTPClient: http_client})
	if err != nil {
		fmt.Printf("Error creating InfluxDB client: %v\n", err)
		panic(err)
	}
	return client
}

func writePoint(topic string, payload InfluxMessage, client influxdb3.Client) error {
	p := influxdb3.NewPoint(payload.Measurement, payload.Tags, payload.Fields, payload.Time)
	points := []*influxdb3.Point{p}
	err := client.WritePoints(context.Background(), points, influxdb3.WithPrecision(lineprotocol.Millisecond))
	if err != nil {
		return err
	}
	return nil
}
