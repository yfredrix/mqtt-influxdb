package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/paho"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type MockClient struct {
	writeAPI api.WriteAPI
}

func (m *MockClient) WriteAPI(org, bucket string) api.WriteAPI {
	return m.writeAPI
}

func (m *MockClient) Close() {
	// Mock close method
}

type MockWriteAPI struct {
	points []*influxdb2.Point
}

func (m *MockWriteAPI) WritePoint(p *influxdb2.Point) {
	m.points = append(m.points, p)
}

func (m *MockWriteAPI) Flush() {
	// Mock flush method
}

func TestNewHandler(t *testing.T) {
	cfg := config{
		influxURL:   "http://localhost:8086",
		influxToken: "testToken",
		influxOrg:   "testOrg",
	}

	handler := NewHandler(cfg)
	if handler == nil {
		t.Fatalf("expected handler to be created, got nil")
	}

	if handler.organization != cfg.influxOrg {
		t.Errorf("expected organization to be %s, got %s", cfg.influxOrg, handler.organization)
	}

	if handler.client == nil {
		t.Errorf("expected client to be created, got nil")
	}
}

func TestHandlerClose(t *testing.T) {
	mockClient := &MockClient{}
	handler := &handler{
		organization: "testOrg",
		client:       mockClient,
	}

	handler.Close()
	// No assertions needed, just ensure no panic
}

func TestSplitTopic(t *testing.T) {
	tests := []struct {
		topic       string
		expectedErr bool
		mainTopic   string
		subTopic    string
	}{
		{"main/sub", false, "main", "sub"},
		{"invalidTopic", true, "", ""},
	}

	for _, test := range tests {
		mainTopic, subTopic, err := splitTopic(test.topic)
		if test.expectedErr && err == nil {
			t.Errorf("expected error for topic %s, got nil", test.topic)
		}
		if !test.expectedErr && err != nil {
			t.Errorf("did not expect error for topic %s, got %v", test.topic, err)
		}
		if mainTopic != test.mainTopic {
			t.Errorf("expected main topic %s, got %s", test.mainTopic, mainTopic)
		}
		if subTopic != test.subTopic {
			t.Errorf("expected sub topic %s, got %s", test.subTopic, subTopic)
		}
	}
}

func TestHandle(t *testing.T) {
	mockWriteAPI := &MockWriteAPI{}
	mockClient := &MockClient{writeAPI: mockWriteAPI}
	handler := &handler{
		organization: "testOrg",
		client:       mockClient,
	}

	payload := Message{
		Measurement: "testMeasurement",
		Tags:        map[string]string{"tag1": "value1"},
		Fields:      map[string]interface{}{"field1": 10},
		Time:        time.Now(),
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &paho.Publish{
		Topic:   "main/sub",
		Payload: payloadBytes,
	}

	handler.handle(msg)

	if len(mockWriteAPI.points) != 1 {
		t.Fatalf("expected 1 point to be written, got %d", len(mockWriteAPI.points))
	}

	point := mockWriteAPI.points[0]
	if point.Name() != "testMeasurement" {
		t.Errorf("expected measurement name to be 'testMeasurement', got %v", point.Name())
	}

	if point.Tags()["tag1"] != "value1" {
		t.Errorf("expected tag 'tag1' to be 'value1', got %v", point.Tags()["tag1"])
	}

	if point.Fields()["field1"] != 10 {
		t.Errorf("expected field 'field1' to be 10, got %v", point.Fields()["field1"])
	}
}
