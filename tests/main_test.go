package main

import (
	"net/url"
	"testing"

	"github.com/eclipse/paho.golang/paho/session/state"
)

func TestLogger(t *testing.T) {
	l := logger{prefix: "test"}

	// Test Println
	l.Println("message")
	// Test Printf
	l.Printf("formatted %s", "message")
}

func TestCreateClient(t *testing.T) {
	serverURL, _ := url.Parse("mqtt://localhost:1883")
	cfg := config{
		serverURL:     serverURL,
		sessionFolder: "",
		debug:         false,
	}

	sessionState := state.NewInMemory()
	h := &handler{}

	clientCfg := createClient(cfg, sessionState, h)
	if clientCfg.ServerUrls[0].String() != serverURL.String() {
		t.Errorf("expected server URL to be %s, got %s", serverURL.String(), clientCfg.ServerUrls[0].String())
	}
}
