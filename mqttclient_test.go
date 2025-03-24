package main

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	"github.com/eclipse/paho.golang/paho/session/state"
)

func TestLoadTLSConfig(t *testing.T) {
	caFile := "testdata/ca.pem"
	clientFile := "testdata/client.pem"
	keyFile := "testdata/client-key.pem"

	// Create test files
	os.WriteFile(caFile, []byte("test-ca"), 0644)
	os.WriteFile(clientFile, []byte("test-client-cert"), 0644)
	os.WriteFile(keyFile, []byte("test-client-key"), 0644)
	defer os.Remove(caFile)
	defer os.Remove(clientFile)
	defer os.Remove(keyFile)

	tlsConfig := loadTLSConfig(caFile, clientFile, keyFile)
	if tlsConfig == nil {
		t.Fatal("expected tlsConfig to be created, got nil")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}

	if tlsConfig.RootCAs == nil {
		t.Errorf("expected RootCAs to be set, got nil")
	}
}

func TestCreateMQTTClient(t *testing.T) {
	serverURL, _ := url.Parse("mqtt://localhost:1883")
	cfg := config{
		serverURL:         serverURL,
		ca:                "testdata/ca.pem",
		cert:              "testdata/client.pem",
		key:               "testdata/client-key.pem",
		keepAlive:         60,
		connectRetryDelay: 5,
		topic:             "test/topic",
		qos:               1,
		clientID:          "testClient",
	}

	sessionState := &state.State{}
	h := &handler{}

	clientCfg := createClient(cfg, sessionState, h)
	if clientCfg.ServerUrls[0].String() != serverURL.String() {
		t.Errorf("expected server URL to be %s, got %s", serverURL.String(), clientCfg.ServerUrls[0].String())
	}

	if clientCfg.KeepAlive != cfg.keepAlive {
		t.Errorf("expected keepAlive to be %d, got %d", cfg.keepAlive, clientCfg.KeepAlive)
	}

	if clientCfg.ReconnectBackoff == nil {
		t.Errorf("expected ReconnectBackoff to be set, got nil")
	}

	if clientCfg.ClientConfig.ClientID != cfg.clientID {
		t.Errorf("expected clientID to be %s, got %s", cfg.clientID, clientCfg.ClientConfig.ClientID)
	}

	if len(clientCfg.ClientConfig.OnPublishReceived) != 1 {
		t.Errorf("expected 1 OnPublishReceived handler, got %d", len(clientCfg.ClientConfig.OnPublishReceived))
	}
}

func TestOnConnectionUp(t *testing.T) {
	serverURL, _ := url.Parse("mqtt://localhost:1883")
	cfg := config{
		serverURL:         serverURL,
		ca:                "testdata/ca.pem",
		cert:              "testdata/client.pem",
		key:               "testdata/client-key.pem",
		keepAlive:         60,
		connectRetryDelay: 5,
		topic:             "test/topic",
		qos:               1,
		clientID:          "testClient",
	}

	sessionState := &state.State{}
	h := &handler{}

	clientCfg := createClient(cfg, sessionState, h)
	if clientCfg.OnConnectionUp == nil {
		t.Fatal("expected OnConnectionUp to be set, got nil")
	}

	// Simulate connection up
	clientCfg.OnConnectionUp(nil, &paho.Connack{})
}

func TestOnConnectError(t *testing.T) {
	serverURL, _ := url.Parse("mqtt://localhost:1883")
	cfg := config{
		serverURL:         serverURL,
		ca:                "testdata/ca.pem",
		cert:              "testdata/client.pem",
		key:               "testdata/client-key.pem",
		keepAlive:         60,
		connectRetryDelay: 5,
		topic:             "test/topic",
		qos:               1,
		clientID:          "testClient",
	}

	sessionState := &state.State{}
	h := &handler{}

	clientCfg := createClient(cfg, sessionState, h)
	if clientCfg.OnConnectError == nil {
		t.Fatal("expected OnConnectError to be set, got nil")
	}

	// Simulate connection error
	clientCfg.OnConnectError(fmt.Errorf("test error"))
}
