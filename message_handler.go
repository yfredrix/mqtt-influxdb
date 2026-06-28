package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eclipse/paho.golang/paho"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
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

type genericPayloadMessage struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
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

type solarMessage struct {
	Model     string                 `json:"model"`
	Data      map[string]interface{} `json:"data"`
	Timestamp float64                `json:"timestamp"`
	Source    string                 `json:"source"`
}

// solarTagKeys is the explicit allowlist of data fields that should be stored as InfluxDB tags.
var solarTagKeys = map[string]bool{
	"status":           true,
	"model_id":         true,
	"model_length":     true,
	"status_vendor_16": true,
	"status_vendor_32": true,
	"status_code":      true,
}

// classifySolarField maps a data field key to its target InfluxDB bucket.
// Returns "tag" for fields that should be stored as tags instead of measurement fields.
// Returns "" for unknown fields that should be skipped.
func classifySolarField(key string) string {
	if solarTagKeys[key] {
		return "tag"
	}
	switch {
	case strings.HasSuffix(key, "_wh"):
		return "latest_energy"
	case strings.HasSuffix(key, "_w") || strings.HasSuffix(key, "_va") ||
		strings.HasSuffix(key, "_var") || strings.HasSuffix(key, "_pct"):
		return "latest_energy_current"
	case strings.HasPrefix(key, "ac_current_") || strings.HasPrefix(key, "ac_voltage_") ||
		strings.HasPrefix(key, "dc_current_") || strings.HasPrefix(key, "dc_voltage_") ||
		strings.HasSuffix(key, "_hz"):
		return "latest_voltage_current"
	case strings.HasPrefix(key, "temp_"):
		return "sensors"
	default:
		return ""
	}
}

func transformSolarValue(key string, val interface{}) (string, interface{}) {
	switch {
	case strings.HasSuffix(key, "_wh"):
		// convert watt-hour values to kilowatt-hours and rename accordingly
		if num, ok := val.(float64); ok {
			return strings.TrimSuffix(key, "_wh") + "_kwh", num / 1000
		}
	case strings.HasSuffix(key, "_w"):
		// convert watt values to kilowatts and rename accordingly
		if num, ok := val.(float64); ok {
			return strings.TrimSuffix(key, "_w") + "_kw", num / 1000
		}
	}
	return key, val
}

// buildSolarPoints parses a raw solar MQTT payload and returns a map of
// bucket name → InfluxMessage ready for writing. Only buckets with at least
// one field are included in the result.
func buildSolarPoints(payload []byte) (map[string]InfluxMessage, error) {
	var solar solarMessage
	if err := json.Unmarshal(payload, &solar); err != nil {
		return nil, err
	}

	sec := int64(solar.Timestamp)
	nsec := int64((solar.Timestamp - float64(sec)) * 1e9)
	timestamp := time.Unix(sec, nsec)

	tags := map[string]string{
		"model":  solar.Model,
		"source": solar.Source,
	}

	buckets := map[string]map[string]interface{}{
		"latest_energy":          {},
		"latest_energy_current":  {},
		"latest_voltage_current": {},
		"sensors":                {},
	}

	for key, val := range solar.Data {
		if val == nil {
			continue
		}
		bucket := classifySolarField(key)
		switch bucket {
		case "tag":
			tags[key] = fmt.Sprintf("%v", val)
		case "":
			fmt.Printf("Unknown solar field %q, skipping\n", key)
		default:
			newKey, transformValue := transformSolarValue(key, val)
			buckets[bucket][newKey] = transformValue
		}
	}

	result := make(map[string]InfluxMessage)
	for bucket, fields := range buckets {
		if len(fields) == 0 {
			continue
		}
		result[bucket] = InfluxMessage{
			Measurement: solar.Source,
			Tags:        tags,
			Fields:      fields,
			Time:        timestamp,
		}
	}
	return result, nil
}

func buildVictronPoint(topic string, payload []byte) (string, InfluxMessage, error) {
	var victronMessage genericPayloadMessage
	if err := json.Unmarshal(payload, &victronMessage); err != nil {
		return "", InfluxMessage{}, fmt.Errorf("topic %q: %w", topic, err)
	}

	splitTopic := strings.Split(topic, "/")
	if len(splitTopic) < 3 {
		return "", InfluxMessage{}, fmt.Errorf("topic is not in the correct format: %s", topic)
	}

	bucket := splitTopic[0]
	serviceType := splitTopic[2]
	deviceInstance := ""
	if len(splitTopic) > 3 {
		deviceInstance = splitTopic[3]
	}

	fieldKey := "value"
	if len(splitTopic) > 4 {
		fieldKey = strings.Join(splitTopic[4:], "/")
	}

	point := InfluxMessage{
		Measurement: serviceType,
		Tags: map[string]string{
			"vrm_portal_id":   splitTopic[1],
			"device_instance": deviceInstance,
		},
		Fields: map[string]interface{}{
			fieldKey: victronMessage.Value,
		},
		Time: time.UnixMilli(victronMessage.Timestamp),
	}

	return bucket, point, nil
}

func handleSolarMessage(msg *paho.Publish, client influxdb2.Client, organization string) {
	points, err := buildSolarPoints(msg.Payload)
	if err != nil {
		fmt.Printf("Solar message could not be parsed (%s): %s", msg.Payload, err)
		return
	}

	writeAPIs := make(map[string]api.WriteAPI)
	for bucket, influxMsg := range points {
		writeAPI := client.WriteAPI(organization, bucket)
		writeAPIs[bucket] = writeAPI
		p := influxdb2.NewPoint(influxMsg.Measurement, influxMsg.Tags, influxMsg.Fields, influxMsg.Time)
		writeAPI.WritePoint(p)
	}
	for _, writeAPI := range writeAPIs {
		writeAPI.Flush()
	}
}

// handle is called when a message is received
func (o *handler) handle(msg *paho.Publish) {
	if strings.HasPrefix(msg.Topic, "solaredge/") {
		handleSolarMessage(msg, o.client, o.organization)
	} else if strings.Contains(msg.Topic, "p1") {
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
	} else if strings.Contains(msg.Topic, "victron") {
		if !strings.HasSuffix(msg.Topic, "Batteries") {
			bucket, victronInfluxMessage, err := buildVictronPoint(msg.Topic, msg.Payload)
			if err != nil {
				fmt.Printf("Victron message could not be parsed (%s): %s", msg.Payload, err)
				return
			}
			writePoint(bucket, victronInfluxMessage, o.client, o.organization)
		}

	} else {
		fmt.Printf("Unknown topic: %s", msg.Topic)
		return
	}
}
