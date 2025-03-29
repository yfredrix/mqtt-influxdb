package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/paho/session/state"
)

func generateTestCerts(caFile, clientFile, keyFile string) error {
	// Generate CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test CA"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPriv.PublicKey, caPriv)
	if err != nil {
		return err
	}

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	// Write CA certificate to file
	if err := os.WriteFile(caFile, caPEM, 0644); err != nil {
		return err
	}

	// Generate client certificate
	client := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{Organization: []string{"Test Client"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	clientBytes, err := x509.CreateCertificate(rand.Reader, client, ca, &clientPriv.PublicKey, caPriv)
	if err != nil {
		return err
	}

	clientPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientBytes})
	clientPrivKey, err := x509.MarshalECPrivateKey(clientPriv)
	if err != nil {
		return err
	}
	clientPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientPrivKey})

	// Write client certificate and key to files
	if err := os.WriteFile(clientFile, clientPEM, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, clientPrivPEM, 0644); err != nil {
		return err
	}

	return nil
}

func TestLoadTLSConfig(t *testing.T) {
	caFile := "ca.pem"
	clientFile := "client.pem"
	keyFile := "client-key.pem"

	// Generate test certificates
	if err := generateTestCerts(caFile, clientFile, keyFile); err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}
	defer os.Remove(caFile)     //nolint:errcheck
	defer os.Remove(clientFile) //nolint:errcheck
	defer os.Remove(keyFile)    //nolint:errcheck

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
		ca:                "ca1.pem",
		cert:              "client1.pem",
		key:               "client-key1.pem",
		keepAlive:         60,
		connectRetryDelay: 5,
		topic:             "test/topic",
		qos:               1,
		clientID:          "testClient",
	}
	// Generate test certificates
	if err := generateTestCerts(cfg.ca, cfg.cert, cfg.key); err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}
	defer os.Remove(cfg.ca)   //nolint:errcheck
	defer os.Remove(cfg.cert) //nolint:errcheck
	defer os.Remove(cfg.key)  //nolint:errcheck
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

	if clientCfg.ClientConfig.ClientID != cfg.clientID { //nolint:staticcheck
		t.Errorf("expected clientID to be %s, got %s", cfg.clientID, clientCfg.ClientConfig.ClientID) //nolint:staticcheck
	}

	if len(clientCfg.ClientConfig.OnPublishReceived) != 1 { //nolint:staticcheck
		t.Errorf("expected 1 OnPublishReceived handler, got %d", len(clientCfg.ClientConfig.OnPublishReceived))
	}
}
