package main

import (
	"encoding/json"
	"testing"

	"github.com/eclipse/paho.golang/paho"
)

func TestClassifySolarField(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		// Energy bucket
		{"ac_energy_wh", "latest_energy"},
		// Power bucket
		{"ac_power_w", "latest_energy_current"},
		{"dc_power_w", "latest_energy_current"},
		{"ac_apparent_power_va", "latest_energy_current"},
		{"ac_reactive_power_var", "latest_energy_current"},
		{"ac_power_factor_pct", "latest_energy_current"},
		// Voltage/current/frequency bucket
		{"ac_current_a", "latest_voltage_current"},
		{"ac_current_b", "latest_voltage_current"},
		{"ac_current_c", "latest_voltage_current"},
		{"ac_current_total", "latest_voltage_current"},
		{"ac_voltage_ab", "latest_voltage_current"},
		{"ac_voltage_an", "latest_voltage_current"},
		{"ac_frequency_hz", "latest_voltage_current"},
		{"dc_current_a", "latest_voltage_current"},
		{"dc_voltage_v", "latest_voltage_current"},
		// Temperature bucket
		{"temp_sink_c", "sensors"},
		// Tags
		{"model_id", "tag"},
		{"model_length", "tag"},
		{"status", "tag"},
		{"status_vendor_16", "tag"},
		{"status_vendor_32", "tag"},
		{"status_code", "tag"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := classifySolarField(tt.key)
			if got != tt.expected {
				t.Errorf("classifySolarField(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestHandleSolarMessage_ParsesAndRoutesFields(t *testing.T) {
	payload := map[string]interface{}{
		"model": "SE2200H/inverter",
		"data": map[string]interface{}{
			"model_id":              101,
			"model_length":          50,
			"ac_current_a":          4.8,
			"ac_current_b":          nil,
			"ac_power_w":            1154.0,
			"ac_apparent_power_va":  1158.1,
			"ac_reactive_power_var": 97.7,
			"ac_power_factor_pct":   99.64,
			"ac_energy_wh":          6679718.0,
			"dc_current_a":          3.172,
			"dc_voltage_v":          369.3,
			"dc_power_w":            1171.6,
			"ac_frequency_hz":       49.966,
			"ac_voltage_an":         240.6,
			"temp_sink_c":           45.14,
			"status":                "MPPT",
			"status_vendor_16":      0,
			"status_vendor_32":      0,
			"status_code":           4,
		},
		"timestamp": 1779634500.6994448,
		"source":    "SE2200H",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal test payload: %v", err)
	}

	written, err := buildSolarPoints(payloadBytes)
	if err != nil {
		t.Fatalf("buildSolarPoints failed: %v", err)
	}

	// Verify all 4 buckets received data
	for _, bucket := range []string{"latest_energy", "latest_energy_current", "latest_voltage_current", "sensors"} {
		if _, ok := written[bucket]; !ok {
			t.Errorf("expected data written to bucket %q, but got nothing", bucket)
		}
	}

	// Verify specific field routing
	if _, ok := written["latest_energy"].Fields["ac_energy_wh"]; !ok {
		t.Error("expected ac_energy_wh in latest_energy bucket")
	}
	if _, ok := written["latest_energy_current"].Fields["ac_power_w"]; !ok {
		t.Error("expected ac_power_w in latest_energy_current bucket")
	}
	if _, ok := written["latest_voltage_current"].Fields["ac_current_a"]; !ok {
		t.Error("expected ac_current_a in latest_voltage_current bucket")
	}
	if _, ok := written["sensors"].Fields["temp_sink_c"]; !ok {
		t.Error("expected temp_sink_c in sensors bucket")
	}

	// Verify null field (ac_current_b) was skipped
	if _, ok := written["latest_voltage_current"].Fields["ac_current_b"]; ok {
		t.Error("ac_current_b was null, should not be in fields")
	}

	// Verify tags are present with expected values
	if written["latest_energy"].Tags["model"] != "SE2200H/inverter" {
		t.Errorf("expected tag model=SE2200H/inverter, got %q", written["latest_energy"].Tags["model"])
	}
	if written["latest_energy"].Tags["source"] != "SE2200H" {
		t.Errorf("expected tag source=SE2200H, got %q", written["latest_energy"].Tags["source"])
	}
	if written["latest_energy"].Tags["status"] != "MPPT" {
		t.Errorf("expected tag status=MPPT, got %q", written["latest_energy"].Tags["status"])
	}

	// Verify measurement name is the source
	if written["latest_energy"].Measurement != "SE2200H" {
		t.Errorf("expected measurement SE2200H, got %q", written["latest_energy"].Measurement)
	}
}

func TestHandleSolarMessage_InvalidPayload(t *testing.T) {
	// Should not panic on invalid payload
	msg := &paho.Publish{
		Topic:   "solaredge/SE2200H/inverter",
		Payload: []byte("not-valid-json"),
	}
	// handleSolarMessage prints error but must not panic
	handleSolarMessage(msg, nil, "test-org")
}
