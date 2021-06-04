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

// newElectrsConnection is a constructor for ElectrsConnection.
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

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf(
				"something went wrong trying to read response for bitcoin transaction broadcast: [%w]; raw transaction: [%s]",
				err,
				transaction,
			)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"failed to broadcast transaction - status: [%s], payload: [%s]; raw transaction: [%s]",
				resp.Status,
				responseBody,
				transaction,
			)
		}

		logger.Infof(
			"successfully broadcast the bitcoin transaction: [%s]",
			responseBody,
		)
		return nil
	})
}

// VbyteFeeFor25Blocks retrieves the 25-block estimate fee per vbyte on the bitcoin network.
func (e electrsConnection) VbyteFeeFor25Blocks() (int32, error) {
	if e.apiURL == "" {
		return 0, fmt.Errorf("attempted to call VbyteFeeFor25Blocks with no apiURL")
	}

	var vbyteFee int32
	err := utils.DoWithDefaultRetry(e.timeout, func(ctx context.Context) error {
		resp, err := e.client.Get(fmt.Sprintf("%s/fee-estimates", e.apiURL))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error(
					"something went wrong trying to read error response for bitcoin fee estimates: [%w]",
					err,
				)
			}

			return fmt.Errorf(
				"failed to get fee estimates - status: [%s], payload: [%s]",
				resp.Status,
				responseBody,
			)
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
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error(
					"something went wrong trying to read error response for transactions of bitcoin address [%s]: [%w]",
					btcAddress,
					err,
				)
			}
			return fmt.Errorf(
				"something went wrong trying to get information about address [%s] - status: [%s], payload: [%s]",
				btcAddress,
				resp.Status,
				responseBody,
			)
		}

		responses := []interface{}{}
		err = json.NewDecoder(resp.Body).Decode(&responses)
		if err != nil {
			return fmt.Errorf("failed to decode response body: [%w]", err)
		}

		isAddressUnused = len(responses) == 0
		return nil
	})
	if err != nil {
		return false, err
	}
	return isAddressUnused, nil
}
