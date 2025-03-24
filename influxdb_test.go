package main

import (
	"testing"

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

func TestInfluxClient(t *testing.T) {
	cfg := config{
		influxURL:   "http://localhost:8086",
		influxToken: "testToken",
	}

	client := influxClient(cfg)
	if client == nil {
		t.Fatalf("expected client to be created, got nil")
	}
}

func TestWritePoint(t *testing.T) {
	mockWriteAPI := &MockWriteAPI{}
	mockClient := &MockClient{writeAPI: mockWriteAPI}

	payload := Message{
		Measurement: "testMeasurement",
		Tags:        map[string]string{"tag1": "value1"},
		Fields:      map[string]interface{}{"field1": 10},
	}

	writePoint("testTopic", payload, mockClient, "testOrg")

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
