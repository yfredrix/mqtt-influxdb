package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	InfluxDB "github.com/influxdata/influxdb/client/v2"
	envConfig "github.com/kelseyhightower/envconfig"
)

var iclient InfluxDB.Client

type Message struct {
	Device      string
	Type        string
	Temperature float64
	Humidity    float64
	Room        string
}

type Location struct {
	City     string
	Building string
	Room     string
}

type Specification struct {
	MQTTHost     string
	MQTTPort     int
	MQTTUser     string
	MQTTPassword string
	MQTTTopic    string

	InfluxHost        string
	InfluxPort        string
	InfluxDatabase    string
	InfluxUser        string
	InfluxPassword    string
	InfluxMeasurement string
}

var s Specification

func init() {
	err := envConfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	iclient, err = InfluxDB.NewHTTPClient(InfluxDB.HTTPConfig{
		Addr:     "http://" + s.InfluxHost + ":" + s.InfluxPort,
		Username: s.InfluxUser,
		Password: s.InfluxPassword,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	hostname, _ := os.Hostname()
	server := fmt.Sprintf("tcp://%s:%d", s.MQTTHost, s.MQTTPort)

	connOpts := MQTT.NewClientOptions()
	connOpts.AddBroker(server)
	connOpts.SetClientID(hostname)
	connOpts.SetCleanSession(true)
	connOpts.SetUsername(s.MQTTUser)
	connOpts.SetPassword(s.MQTTPassword)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	connOpts.SetTLSConfig(tlsConfig)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(s.MQTTTopic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to ", server)
	}

	<-c
}

// onMessageReceived is triggered by subscription on MQTTTopic (default #)
func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	log.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())

	//verify the MQTT topic
	location, err := verifyTopic(message.Topic())
	if err != nil {
		log.Println(err)
		return
	}

	// write data to InfluxDB
	err = writeToInfluxDB(string(message.Payload()), location)
	if err != nil {
		// just log the error but keep going
		log.Println("Error while writing to InfluxDB: ", err)
	}
}

// verifyTopic checks the MQTT topic conforms to airq/city/building/room
func verifyTopic(topic string) (*Location, error) {
	location := &Location{}
	items := strings.Split(topic, "/")
	if len(items) != 4 {
		return nil, errors.New("MQTT topic requires 4 sections: airq, city, building, room")
	}

	location.City = items[1]
	location.Building = items[2]
	location.Room = items[3]

	if items[0] != "airq" {
		return nil, errors.New("MQTT topic needs to start with airq")
	}

	if location.City == "" || location.Building == "" || location.Room == "" {
		return nil, errors.New("MQTT topic needs to have a city, building and room")
	}

	return location, nil
}
func writeToInfluxDB(message string, location *Location) error {
	// decode message (json) in m which is Message struct
	// json in form of: {"device":"deviceX", "temperature": 23.3,"humidity": 50.2,"co2": 200}
	var m Message
	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		return err
	}

	m.Room = location.Room

	log.Printf("Writing message %s from room %s to database %s\n", message, location.Room, s.InfluxDatabase)

	// Create a new point batch
	bp, err := InfluxDB.NewBatchPoints(InfluxDB.BatchPointsConfig{
		Database:  s.InfluxDatabase,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{"device": m.Device, "type": m.Type, "room": m.Room}
	fields := map[string]interface{}{
		"temperature": m.Temperature,
		"humidity":    m.Humidity,
	}

	pt, err := InfluxDB.NewPoint(s.InfluxMeasurement, tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := iclient.Write(bp); err != nil {
		return err
	}

	return nil
}
