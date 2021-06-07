package bitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-ecdsa/pkg/utils"
)

var logger = log.Logger("bitcoin")

const (
	defaultTimeout = 2 * time.Minute
)

type httpClient interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// electrsConnection exposes a native API for interacting with an electrs http API.
type electrsConnection struct {
	apiURL  string
	client  httpClient
	timeout time.Duration
}

// Connect is a constructor for electrsConnection.
func Connect(apiURL string) Handle {
	return &electrsConnection{
		apiURL:  apiURL,
		client:  http.DefaultClient,
		timeout: defaultTimeout,
	}
}

func (e *electrsConnection) setClient(client httpClient) {
	e.client = client
}

// Broadcast broadcasts a transaction the configured bitcoin network.
func (e electrsConnection) Broadcast(transaction string) error {
	if e.apiURL == "" {
		return fmt.Errorf("attempted to call Broadcast with no apiURL")
	}
	return utils.DoWithDefaultRetry(e.timeout, func(ctx context.Context) error {
		resp, err := e.client.Post(fmt.Sprintf("%s/tx", e.apiURL), "text/plain", strings.NewReader(transaction))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("something went wrong with broadcast: [%s], transaction: [%s]", resp.Status, transaction)
		}
		transactionIDBuffer := new(strings.Builder)
		bytesCopied, err := io.Copy(transactionIDBuffer, resp.Body)
		if err != nil {
			return fmt.Errorf("something went wrong reading the electrs response body: [%v]", err)
		}
		if bytesCopied == 0 {
			return fmt.Errorf("something went wrong reading the electrs response body: 0 bytes copied")
		}
		logger.Infof("successfully broadcast the bitcoin transaction: %s", transactionIDBuffer.String())
		return nil
	})
}

// VbyteFeeFor25Blocks retrieves the 25-block estimate fee per vbyte on the bitcoin network.
func (e electrsConnection) VbyteFeeFor25Blocks() (int32, error) {
	if e.apiURL == "" {
		return 0, fmt.Errorf("attempted to call VbyteFee with no apiURL")
	}
	var vbyteFee int32
	err := utils.DoWithDefaultRetry(e.timeout, func(ctx context.Context) error {
		resp, err := e.client.Get(fmt.Sprintf("%s/fee-estimates", e.apiURL))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("something went wrong retreiving the vbyte fee: [%s]", resp.Status)
		}

		var fees map[string]float32
		err = json.NewDecoder(resp.Body).Decode(&fees)
		if err != nil {
			return fmt.Errorf("something went wrong decoding the vbyte fees: [%v]", err)
		}
		fee, ok := fees["25"]
		if !ok {
			fee = 0
		}
		logger.Info("retrieved a vbyte fee of [%d]", fee)
		vbyteFee = int32(fee)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return vbyteFee, nil
}

// IsAddressUnused returns true if and only if the supplied bitcoin address has
// no recorded transactions.
func (e electrsConnection) IsAddressUnused(btcAddress string) (bool, error) {
	if e.apiURL == "" {
		return false, fmt.Errorf("attempted to call IsAddressUnused with no apiURL")
	}
	isAddressUnused := false
	err := utils.DoWithDefaultRetry(e.timeout, func(ctx context.Context) error {
		resp, err := e.client.Get(fmt.Sprintf("%s/address/%s/txs", e.apiURL, btcAddress))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			transactionIDBuffer := new(strings.Builder)
			_, err = io.Copy(transactionIDBuffer, resp.Body)
			if err != nil {
				logger.Error("something went wrong trying to unmarshal the error response for address %s", btcAddress)
			}
			return fmt.Errorf(
				"something went wrong trying to get information about address %s - status: [%s], payload: [%s]",
				btcAddress,
				resp.Status,
				transactionIDBuffer.String(),
			)
		}

		responses := []interface{}{}
		err = json.NewDecoder(resp.Body).Decode(&responses)
		if err != nil {
			return err
		}

		isAddressUnused = len(responses) == 0
		return nil
	})
	if err != nil {
		return false, err
	}
	return isAddressUnused, nil
}
