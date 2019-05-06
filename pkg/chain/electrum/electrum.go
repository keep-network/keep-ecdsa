// Package electrum contains implementation of the chain interface communicating
// with electrum network.
package electrum

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	goElectrum "github.com/keep-network/go-electrum/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/chain"
)

type electrum struct {
	electrumServer *goElectrum.Server
}

// Config contains configuration of Electrum network.
type Config struct {
	// Connection details of the Electrum server.
	ServerHost, ServerPort string
}

// PublishTransaction sends a raw transaction provided as a hexadecimal string
// to the electrum server. It returns a transaction hash as a hexadecimal string.
func (e *electrum) PublishTransaction(rawTx string) (string, error) {
	return e.electrumServer.BroadcastTransaction(rawTx)
}

// Connect establishes connection to the Electrum Server defined in a provided
// config. The server is expected to be connected to a specific network.
func Connect(config *Config) (chain.Interface, error) {
	serverAddress := strings.Join([]string{config.ServerHost, config.ServerPort}, ":")

	// TODO: Ignore server certificates is a temporary solution for development.
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	server := goElectrum.NewServer()

	// TODO: For development we support only SSL connections because there are
	// more SSL than TCP electrum servers available for BTC testnet.
	if err := server.ConnectSSL(serverAddress, tlsConfig); err != nil {
		return nil, fmt.Errorf("connecting to %s failed [%s]", serverAddress, err)
	}

	serverVersion, protocolVersion, err := server.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("cannot get server version [%s]", err)
	}
	fmt.Printf(
		"Connected to Electrum Server.\nServer version: %s [Protocol %s]\n",
		serverVersion,
		protocolVersion,
	)

	// Keep connection alive.
	go func() {
		for {
			if err := server.Ping(); err != nil {
				fmt.Printf("ping failed [%s]", err)
			}
			time.Sleep(60 * time.Second)
		}
	}()

	return &electrum{electrumServer: server}, nil
}
