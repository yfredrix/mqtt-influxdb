package main

import "testing"

func TestBuildVictronPoint_ExampleMessages(t *testing.T) {
	tests := []struct {
		name           string
		topic          string
		payload        []byte
		expectedBucket string
		expectedMetric string
		expectedPortal string
		expectedDevice string
		expectedField  string
		expectedValue  float64
		expectedMillis int64
	}{
		{
			name:           "grid power message",
			topic:          "victron/a7f3c19de82b/grid/40/Ac/L3/Power",
			payload:        []byte(`{"value": -1393, "timestamp": 1782637540236}`),
			expectedBucket: "victron",
			expectedMetric: "grid",
			expectedPortal: "a7f3c19de82b",
			expectedDevice: "40",
			expectedField:  "Ac/L3/Power",
			expectedValue:  -1393,
			expectedMillis: 1782637540236,
		},
		{
			name:           "system pv output power message",
			topic:          "victron/f29b4d80a6ce/system/0/Ac/PvOnOutput/L1/Power",
			payload:        []byte(`{"value": 1527.6, "timestamp": 1782637540383}`),
			expectedBucket: "victron",
			expectedMetric: "system",
			expectedPortal: "f29b4d80a6ce",
			expectedDevice: "0",
			expectedField:  "Ac/PvOnOutput/L1/Power",
			expectedValue:  1527.6,
			expectedMillis: 1782637540383,
		},
		{
			name:           "Max Discharge Power Message",
			topic:          "victron/a7f3c19de82b/hub4/0/MaxDischargePower",
			payload:        []byte(`{"value": 4688.9998626709, "timestamp": 1782637544452}`),
			expectedBucket: "victron",
			expectedMetric: "hub4",
			expectedPortal: "a7f3c19de82b",
			expectedDevice: "0",
			expectedField:  "MaxDischargePower",
			expectedValue:  4688.9998626709,
			expectedMillis: 1782637544452,
		},
		{
			name:           "heartbeat message with short topic",
			topic:          "victron/a7f3c19de82b/heartbeat",
			payload:        []byte(`{"value": 1782637540, "timestamp": 1782637540438}`),
			expectedBucket: "victron",
			expectedMetric: "heartbeat",
			expectedPortal: "a7f3c19de82b",
			expectedDevice: "",
			expectedField:  "value",
			expectedValue:  1782637540,
			expectedMillis: 1782637540438,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, point, err := buildVictronPoint(tt.topic, tt.payload)
			if err != nil {
				t.Fatalf("buildVictronPoint returned error: %v", err)
			}

			if bucket != tt.expectedBucket {
				t.Errorf("expected bucket %q, got %q", tt.expectedBucket, bucket)
			}
			if point.Measurement != tt.expectedMetric {
				t.Errorf("expected measurement %q, got %q", tt.expectedMetric, point.Measurement)
			}
			if point.Tags["VRM_Portal_ID"] != tt.expectedPortal {
				t.Errorf("expected VRM_Portal_ID %q, got %q", tt.expectedPortal, point.Tags["VRM_Portal_ID"])
			}
			if point.Tags["Device_Instance"] != tt.expectedDevice {
				t.Errorf("expected Device_Instance %q, got %q", tt.expectedDevice, point.Tags["Device_Instance"])
			}

			gotValue, ok := point.Fields[tt.expectedField]
			if !ok {
				t.Fatalf("expected field key %q to exist", tt.expectedField)
			}
			if gotValue != tt.expectedValue {
				t.Errorf("expected field value %v, got %v", tt.expectedValue, gotValue)
			}
			if point.Time.UnixMilli() != tt.expectedMillis {
				t.Errorf("expected timestamp %d, got %d", tt.expectedMillis, point.Time.UnixMilli())
			}
		})
	}
}

func TestBuildVictronPoint_InvalidInput(t *testing.T) {
	if _, _, err := buildVictronPoint("victron/too-short", []byte(`{"value": 1, "timestamp": 2}`)); err == nil {
		t.Fatal("expected error for malformed topic")
	}

	if _, _, err := buildVictronPoint("victron/a/grid/1/x", []byte(`not-json`)); err == nil {
		t.Fatal("expected error for invalid payload")
	}
}
