package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/eclipse/paho.golang/paho/session/state"
)

func loadTLSConfig(caFile string, clientFile string, keyFile string) (*tls.Config, error) {
	// load tls config
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = false
	if caFile != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
			err := fmt.Errorf("missing system cert pool, using empty pool")
			return nil, err
		}
		ca, err := os.ReadFile(caFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
			log.Println("No certs appended, using system certs only")
		}

		// Import client certificate/key pair
		cert, err := tls.LoadX509KeyPair(clientFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = rootCAs
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}

func createClient(cfg config, sessionState *state.State, h *handler) autopaho.ClientConfig {
	TlsCfg, err := loadTLSConfig(cfg.ca, cfg.cert, cfg.key)
	if err != nil {
		panic(err)
	}
	// Create a handler that will deal with incoming messages
	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{cfg.serverURL},
		TlsCfg:                        TlsCfg,
		KeepAlive:                     cfg.keepAlive,
		CleanStartOnInitialConnection: false, // the default
		SessionExpiryInterval:         60,    // Session remains live 60 seconds after disconnect
		ReconnectBackoff:              autopaho.NewConstantBackoff(cfg.connectRetryDelay),
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: cfg.topic, QoS: cfg.qos},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
				return
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError: func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.clientID,
			Session:  sessionState,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					h.handle(pr.Packet)
					return true, nil
				}},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	return cliCfg
}
