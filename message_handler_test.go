package main

import (
	"testing"

	"github.com/eclipse/paho.golang/paho"
)

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
			name:           "four-part topic defaults field key to value",
			topic:          "victron/a7f3c19de82b/hub4/0",
			payload:        []byte(`{"value": 1250.5, "timestamp": 1782637540999}`),
			expectedBucket: "victron",
			expectedMetric: "hub4",
			expectedPortal: "a7f3c19de82b",
			expectedDevice: "0",
			expectedField:  "value",
			expectedValue:  1250.5,
			expectedMillis: 1782637540999,
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
			if point.Tags["vrm_portal_id"] != tt.expectedPortal {
				t.Errorf("expected vrm_portal_id %q, got %q", tt.expectedPortal, point.Tags["vrm_portal_id"])
			}
			if point.Tags["device_instance"] != tt.expectedDevice {
				t.Errorf("expected device_instance %q, got %q", tt.expectedDevice, point.Tags["device_instance"])
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

func TestHandle_SkipsVictronTopics(t *testing.T) {
	h := &handler{organization: "test-org", client: nil}
	msg := &paho.Publish{
		Topic: "victron/f29b4d80a6ce/system/0/Batteries",
		Payload: []byte(`{
			"value": [
				{
					"voltage": 50.09000015258789,
					"temperature": 29.5,
					"state": 1,
					"soc": 53,
					"power": 616,
					"name": "Pylontech battery",
					"instance": 512,
					"id": "com.victronenergy.battery.socketcan_vecan1",
					"current": 12.300000190734863,
					"active_battery_service": true
				}
			],
			"timestamp": 1782637542140
		}`),
	}
	msg2 := &paho.Publish{
		Topic: "victron/f29b4d80a6ce/system/0/Network/Services",
		Payload: []byte(`{
			"value": [{
			"ethernet": {
    "Wired": {
      "Address": "10.0.0.10",
      "Gateway": "10.0.0.1",
      "Mac": "AA:BB:CC:DD:EE:01",
      "Method": "dhcp",
      "Nameservers": ["10.0.0.1"],
      "Netmask": "255.255.255.0",
      "Service": "/net/connman/service/ethernet_aabbccddee01_cable",
      "State": "ready"
    }
  },
  "wifi": {
    "<hidden>": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_hidden_managed_psk",
      "State": "idle",
      "Strength": 30
    },
    "WiFi_1": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_ssid01_managed_psk",
      "State": "idle",
      "Strength": 32
    },
    "WiFi_2": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_ssid02_managed_psk",
      "State": "idle",
      "Strength": 59
    },
    "WiFi_3": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_ssid03_managed_psk",
      "State": "idle",
      "Strength": 59
    },
    "WiFi_4": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_ssid04_managed_psk",
      "State": "idle",
      "Strength": 37
    },
    "WiFi_5": {
      "Address": "",
      "Favorite": "no",
      "Gateway": "",
      "Mac": "AA:BB:CC:DD:EE:02",
      "Method": "",
      "Nameservers": [],
      "Netmask": "",
      "Secured": "yes",
      "Service": "/net/connman/service/wifi_aabbccddee02_ssid05_managed_psk",
      "State": "idle",
      "Strength": 29
    }
  }
}]}`),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected skipped topics to not write, but handle panicked: %v", r)
		}
	}()

	h.handle(msg)
	h.handle(msg2)
}
