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
		return "", "", fmt.Errorf("Topic is not in the correct format: %s", topic)
	}
	mainTopic := topicSplit[0]
	subTopic := topicSplit[1]
	return mainTopic, subTopic, nil
}

// Message is used for marshalling/unmarshalling the JSON message (just a count)
type Message struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Time        time.Time              `json:"time"`
}

// handle is called when a message is received
func (o *handler) handle(msg *paho.Publish) {
	// We extract the json structure from the payload
	var p1Message Message
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
}
