package main

// Connect to the server, subscribe, and write messages received to a file

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho/session/state"
	storefile "github.com/eclipse/paho.golang/paho/store/file"
)

func main() {
	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}

	// Create a handler that will deal with incoming messages
	h := NewHandler(cfg)
	defer h.Close()

	var sessionState *state.State
	var cliCfg autopaho.ClientConfig
	if len(cfg.sessionFolder) == 0 {
		sessionState = state.NewInMemory()
	} else {
		cliState, err := storefile.New(cfg.sessionFolder, "subdemo_cli_", ".pkt")
		if err != nil {
			panic(err)
		}
		srvState, err := storefile.New(cfg.sessionFolder, "subdemo_srv_", ".pkt")
		if err != nil {
			panic(err)
		}
		sessionState = state.New(cliState, srvState)
	}

	cliCfg = createClient(cfg, sessionState, h)

	if cfg.debug {
		cliCfg.Debug = logger{prefix: "autoPaho"}
		cliCfg.PahoDebug = logger{prefix: "paho"}
	}

	//
	// Connect to the server
	//
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		panic(err)
	}

	// Messages will be handled through the callback so we really just need to wait until a shutdown
	// is requested
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")

	// We could cancel the context at this point but will call Disconnect instead (this waits for autopaho to shutdown)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = cm.Disconnect(ctx)

	fmt.Println("shutdown complete")
}

// logger implements the paho.Logger interface
type logger struct {
	prefix string
}

// Println is the library provided NOOPLogger's
// implementation of the required interface function()
func (l logger) Println(v ...interface{}) {
	fmt.Println(append([]interface{}{l.prefix + ":"}, v...)...)
}

// Printf is the library provided NOOPLogger's
// implementation of the required interface function(){}
func (l logger) Printf(format string, v ...interface{}) {
	if len(format) > 0 && format[len(format)-1] != '\n' {
		format = format + "\n" // some log calls in paho do not add \n
	}
	fmt.Printf(l.prefix+":"+format, v...)
}
