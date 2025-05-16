package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eclipse/paho.golang/paho"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// handler is a simple struct that provides a function to be called when a message is received. The message is parsed
// and the count followed by the raw message is written to the file (this makes it easier to sort the file)
type handler struct {
	organization string
	client       influxdb2.Client
}

// NewHandler creates a new output handler and opens the output file (if applicable)
func NewHandler(cfg config) *handler {

	return &handler{
		organization: cfg.influxOrg,
		client:       influxClient(cfg),
	}
}

// Close closes the influxDB client
func (o *handler) Close() {
	o.client.Close()
}

func splitTopic(topic string) (string, string, error) {
	topicSplit := strings.Split(topic, "/")
	if len(topicSplit) != 2 {
		err := fmt.Errorf("topic is not in the correct format: %s", topic)
		return "", "", err
	}
	mainTopic := topicSplit[0]
	subTopic := topicSplit[1]
	return mainTopic, subTopic, nil
}

type sensorMessage struct {
	Unit      string    `json:"unit"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type InfluxMessage struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Time        time.Time              `json:"time"`
}

func toInfluxMessage(measurement string, location string, sensorId string, message sensorMessage) InfluxMessage {
	return InfluxMessage{
		Measurement: measurement,
		Tags: map[string]string{
			"unit":     message.Unit,
			"location": location,
		},
		Fields: map[string]interface{}{
			sensorId: message.Value,
		},
		Time: message.Timestamp,
	}
}

// handle is called when a message is received
func (o *handler) handle(msg *paho.Publish) {
	if strings.Contains(msg.Topic, "p1") {
		var p1Message InfluxMessage
		err := json.Unmarshal(msg.Payload, &p1Message)
		if err != nil {
			fmt.Printf("Message could not be parsed (%s): %s", msg.Payload, err)
		}
		_, subTopic, err := splitTopic(msg.Topic)
		if err != nil {
			fmt.Printf("Error splitting topic: %s", err)
			return
		}
		writePoint(subTopic, p1Message, o.client, o.organization)
	} else if strings.Contains(msg.Topic, "sensors") {
		var sensorMessage sensorMessage
		err := json.Unmarshal(msg.Payload, &sensorMessage)
		if err != nil {
			fmt.Printf("Message could not be parsed (%s): %s", msg.Payload, err)
		}
		if sensorMessage.Timestamp.IsZero() {
			sensorMessage.Timestamp = time.Now()
		}

		splittedTopic := strings.Split(msg.Topic, "/")
		if len(splittedTopic) != 4 {
			fmt.Printf("Topic is not in the correct format: %s", msg.Topic)
			return
		}
		bucket, measurement, location, sensorId := splittedTopic[0], splittedTopic[1], splittedTopic[2], splittedTopic[3]

		sensorInfluxMessage := toInfluxMessage(measurement, location, sensorId, sensorMessage)

		writePoint(bucket, sensorInfluxMessage, o.client, o.organization)
	} else {
		fmt.Printf("Unknown topic: %s", msg.Topic)
		return
	}
}
