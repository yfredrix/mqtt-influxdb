package main

import (
	"os"
	"testing"
	"time"
)

func setEnv(key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		panic(err)
	}
}

func unsetEnv(key string) {
	err := os.Unsetenv(key)
	if err != nil {
		panic(err)
	}
}

func TestGetConfig(t *testing.T) {
	setEnv(envServerURL, "http://localhost:1883")
	setEnv(envClientID, "testClient")
	setEnv(envTopic, "test/topic")
	setEnv(envQos, "1")
	setEnv(caFile, "path/to/ca.pem")
	setEnv(clientFile, "path/to/client.pem")
	setEnv(keyFile, "path/to/key.pem")
	setEnv(envKeepAlive, "60")
	setEnv(envConnectRetryDelay, "1000")
	setEnv(envSessionFolder, "/tmp/session")
	setEnv(envDebug, "true")
	setEnv(influxURL, "http://localhost:8086")
	setEnv(influxToken, "testToken")
	setEnv(influxOrg, "testOrg")

	defer func() {
		unsetEnv(envServerURL)
		unsetEnv(envClientID)
		unsetEnv(envTopic)
		unsetEnv(envQos)
		unsetEnv(caFile)
		unsetEnv(clientFile)
		unsetEnv(keyFile)
		unsetEnv(envKeepAlive)
		unsetEnv(envConnectRetryDelay)
		unsetEnv(envSessionFolder)
		unsetEnv(envDebug)
		unsetEnv(influxURL)
		unsetEnv(influxToken)
		unsetEnv(influxOrg)
	}()

	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.serverURL.String() != "http://localhost:1883" {
		t.Errorf("expected serverURL to be 'http://localhost:1883', got %v", cfg.serverURL)
	}
	if cfg.clientID != "testClient" {
		t.Errorf("expected clientID to be 'testClient', got %v", cfg.clientID)
	}
	if cfg.topic != "test/topic" {
		t.Errorf("expected topic to be 'test/topic', got %v", cfg.topic)
	}
	if cfg.qos != 1 {
		t.Errorf("expected qos to be 1, got %v", cfg.qos)
	}
	if cfg.ca != "path/to/ca.pem" {
		t.Errorf("expected ca to be 'path/to/ca.pem', got %v", cfg.ca)
	}
	if cfg.cert != "path/to/client.pem" {
		t.Errorf("expected cert to be 'path/to/client.pem', got %v", cfg.cert)
	}
	if cfg.key != "path/to/key.pem" {
		t.Errorf("expected key to be 'path/to/key.pem', got %v", cfg.key)
	}
	if cfg.keepAlive != 60 {
		t.Errorf("expected keepAlive to be 60, got %v", cfg.keepAlive)
	}
	if cfg.connectRetryDelay != 1000*time.Millisecond {
		t.Errorf("expected connectRetryDelay to be 1000ms, got %v", cfg.connectRetryDelay)
	}
	if cfg.sessionFolder != "/tmp/session" {
		t.Errorf("expected sessionFolder to be '/tmp/session', got %v", cfg.sessionFolder)
	}
	if cfg.debug != true {
		t.Errorf("expected debug to be true, got %v", cfg.debug)
	}
	if cfg.influxURL != "http://localhost:8086" {
		t.Errorf("expected influxURL to be 'http://localhost:8086', got %v", cfg.influxURL)
	}
	if cfg.influxToken != "testToken" {
		t.Errorf("expected influxToken to be 'testToken', got %v", cfg.influxToken)
	}
	if cfg.influxOrg != "testOrg" {
		t.Errorf("expected influxOrg to be 'testOrg', got %v", cfg.influxOrg)
	}
}

func TestStringFromEnv(t *testing.T) {
	setEnv("TEST_STRING", "testValue")
	defer unsetEnv("TEST_STRING")

	value, err := stringFromEnv("TEST_STRING")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != "testValue" {
		t.Errorf("expected 'testValue', got %v", value)
	}
}

func TestIntFromEnv(t *testing.T) {
	setEnv("TEST_INT", "123")
	defer unsetEnv("TEST_INT")

	value, err := intFromEnv("TEST_INT", 16)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 123 {
		t.Errorf("expected 123, got %v", value)
	}
}

func TestMilliSecondsFromEnv(t *testing.T) {
	setEnv("TEST_MS", "1000")
	defer unsetEnv("TEST_MS")

	value, err := milliSecondsFromEnv("TEST_MS")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 1000*time.Millisecond {
		t.Errorf("expected 1000ms, got %v", value)
	}
}

func TestBooleanFromEnv(t *testing.T) {
	setEnv("TEST_BOOL", "true")
	defer unsetEnv("TEST_BOOL")

	value, err := booleanFromEnv("TEST_BOOL")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != true {
		t.Errorf("expected true, got %v", value)
	}
}
func TestIntFromEnvInvalidValue(t *testing.T) {
	setEnv("TEST_INT", "not_a_number")
	defer unsetEnv("TEST_INT")

	_, err := intFromEnv("TEST_INT", 16)
	if err == nil {
		t.Fatal("expected error for invalid integer value")
	}
}

func TestIntFromEnvMissing(t *testing.T) {
	_, err := intFromEnv("NONEXISTENT_VAR", 16)
	if err == nil {
		t.Fatal("expected error for missing environmental variable")
	}
}

func TestIntFromEnvEmpty(t *testing.T) {
	setEnv("TEST_INT", "")
	defer unsetEnv("TEST_INT")

	_, err := intFromEnv("TEST_INT", 16)
	if err == nil {
		t.Fatal("expected error for empty environmental variable")
	}
}

func TestIntFromEnvNegative(t *testing.T) {
	setEnv("TEST_INT", "-123")
	defer unsetEnv("TEST_INT")

	_, err := intFromEnv("TEST_INT", 16)
	if err == nil {
		t.Fatal("expected error for negative integer")
	}
}

func TestInfluxWriteBatchSizeDefault(t *testing.T) {
	setEnv(envServerURL, "http://localhost:1883")
	setEnv(envClientID, "testClient")
	setEnv(envTopic, "test/topic")
	setEnv(envQos, "1")
	setEnv(caFile, "path/to/ca.pem")
	setEnv(clientFile, "path/to/client.pem")
	setEnv(keyFile, "path/to/key.pem")
	setEnv(envKeepAlive, "60")
	setEnv(envConnectRetryDelay, "1000")
	setEnv(influxURL, "http://localhost:8086")
	setEnv(influxToken, "testToken")
	setEnv(influxOrg, "testOrg")
	defer func() {
		unsetEnv(envServerURL)
		unsetEnv(envClientID)
		unsetEnv(envTopic)
		unsetEnv(envQos)
		unsetEnv(caFile)
		unsetEnv(clientFile)
		unsetEnv(keyFile)
		unsetEnv(envKeepAlive)
		unsetEnv(envConnectRetryDelay)
		unsetEnv(influxURL)
		unsetEnv(influxToken)
		unsetEnv(influxOrg)
		unsetEnv(envInfluxWriteBatchSize)
		unsetEnv(envInfluxFlushInterval)
	}()

	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.influxWriteBatchSize != 5000 {
		t.Errorf("expected default influxWriteBatchSize 5000, got %d", cfg.influxWriteBatchSize)
	}
	if cfg.influxFlushInterval != 1000*time.Millisecond {
		t.Errorf("expected default influxFlushInterval 1000ms, got %v", cfg.influxFlushInterval)
	}
}

func TestInfluxWriteBatchSizeZeroRejected(t *testing.T) {
	setEnv(envInfluxWriteBatchSize, "0")
	defer unsetEnv(envInfluxWriteBatchSize)

	_, err := intFromEnvWithDefault(envInfluxWriteBatchSize, 5000, 32)
	if err != nil {
		t.Fatalf("intFromEnvWithDefault returned unexpected error: %v", err)
	}
	// Validate that getConfig rejects a zero batch size end-to-end
	setEnv(envServerURL, "http://localhost:1883")
	setEnv(envClientID, "testClient")
	setEnv(envTopic, "test/topic")
	setEnv(envQos, "1")
	setEnv(caFile, "path/to/ca.pem")
	setEnv(clientFile, "path/to/client.pem")
	setEnv(keyFile, "path/to/key.pem")
	setEnv(envKeepAlive, "60")
	setEnv(envConnectRetryDelay, "1000")
	setEnv(influxURL, "http://localhost:8086")
	setEnv(influxToken, "testToken")
	setEnv(influxOrg, "testOrg")
	defer func() {
		unsetEnv(envServerURL)
		unsetEnv(envClientID)
		unsetEnv(envTopic)
		unsetEnv(envQos)
		unsetEnv(caFile)
		unsetEnv(clientFile)
		unsetEnv(keyFile)
		unsetEnv(envKeepAlive)
		unsetEnv(envConnectRetryDelay)
		unsetEnv(influxURL)
		unsetEnv(influxToken)
		unsetEnv(influxOrg)
	}()

	_, err = getConfig()
	if err == nil {
		t.Fatal("expected error when INFLUXDB_WRITE_BATCH_SIZE=0, got nil")
	}
}

func TestInfluxWriteBatchSizeInvalidRejected(t *testing.T) {
	_, err := intFromEnvWithDefault(envInfluxWriteBatchSize, 5000, 32)
	if err != nil {
		t.Fatalf("unexpected error with missing env var: %v", err)
	}

	setEnv(envInfluxWriteBatchSize, "not-a-number")
	defer unsetEnv(envInfluxWriteBatchSize)

	_, err = intFromEnvWithDefault(envInfluxWriteBatchSize, 5000, 32)
	if err == nil {
		t.Fatal("expected error for non-integer INFLUXDB_WRITE_BATCH_SIZE")
	}
}

func TestInfluxFlushIntervalZeroRejected(t *testing.T) {
	setEnv(envInfluxFlushInterval, "0")
	defer unsetEnv(envInfluxFlushInterval)

	_, err := milliSecondsFromEnvWithDefault(envInfluxFlushInterval, 1000)
	if err == nil {
		t.Fatal("expected error when INFLUXDB_FLUSH_INTERVAL_MS=0")
	}
}

func TestInfluxFlushIntervalNegativeRejected(t *testing.T) {
	setEnv(envInfluxFlushInterval, "-500")
	defer unsetEnv(envInfluxFlushInterval)

	_, err := milliSecondsFromEnvWithDefault(envInfluxFlushInterval, 1000)
	if err == nil {
		t.Fatal("expected error when INFLUXDB_FLUSH_INTERVAL_MS is negative")
	}
}

func TestInfluxFlushIntervalInvalidRejected(t *testing.T) {
	setEnv(envInfluxFlushInterval, "not-a-number")
	defer unsetEnv(envInfluxFlushInterval)

	_, err := milliSecondsFromEnvWithDefault(envInfluxFlushInterval, 1000)
	if err == nil {
		t.Fatal("expected error for non-integer INFLUXDB_FLUSH_INTERVAL_MS")
	}
}

func TestInfluxFlushIntervalCustomValue(t *testing.T) {
	setEnv(envInfluxFlushInterval, "2500")
	defer unsetEnv(envInfluxFlushInterval)

	d, err := milliSecondsFromEnvWithDefault(envInfluxFlushInterval, 1000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d != 2500*time.Millisecond {
		t.Errorf("expected 2500ms, got %v", d)
	}
}
