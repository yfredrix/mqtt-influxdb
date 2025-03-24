package main

import (
	"net/url"
	"testing"

	"github.com/eclipse/paho.golang/paho/session/state"
	storefile "github.com/eclipse/paho.golang/paho/store/file"
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

func TestSessionState(t *testing.T) {
	cfg := config{
		sessionFolder: "testdata/session",
	}

	cliState, err := storefile.New(cfg.sessionFolder, "subdemo_cli_", ".pkt")
	if err != nil {
		t.Fatalf("failed to create client state: %v", err)
	}
	srvState, err := storefile.New(cfg.sessionFolder, "subdemo_srv_", ".pkt")
	if err != nil {
		t.Fatalf("failed to create server state: %v", err)
	}
	sessionState := state.New(cliState, srvState)
	if sessionState == nil {
		t.Fatal("expected sessionState to be created, got nil")
	}
}
