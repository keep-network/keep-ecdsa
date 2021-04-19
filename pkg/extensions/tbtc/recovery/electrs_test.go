package recovery

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type mockClient struct {
	mockGet  func(url string) (*http.Response, error)
	mockPost func(url string, contentType string, reader io.Reader) (*http.Response, error)
}

func (m mockClient) Get(url string) (*http.Response, error) {
	return m.mockGet(url)
}

func (m mockClient) Post(url string, contentType string, reader io.Reader) (*http.Response, error) {
	return m.mockPost(url, contentType, reader)
}

func TestElectrsConnection_broadcast(t *testing.T) {
	apiURL := "example.org/api"
	electrs := NewElectrsConnection(apiURL)
	transaction := "01000000000101ba84a592005742406bd1d6683e3a894c7ab13385bd437ff7bd7c74929bf141320000000000000000000309cc320000000000160014a0aedee089b0cfa34e1e29c2dd2e618b19e8b95309cc320000000000160014f8c4e8695f8c2e0f598f8f00c2c4a83b17b0c4fa09cc320000000000160014fada4235022b32a31f97adbc954e6a7bbb7b32ba024830450221008dd10d4f331a61c2afe948dec6f900b29996c29262a79fa4d72acacd0c19497a022063229d6751c47e3e9b67e567bd98500b66630a09ef8515377a18d0479135c84f01210329fb706ee25a944362c4a53a5b4fa6f47201354d567e753c3998f15a36996b8100000000"
	electrs.setClient(mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return nil, nil
		},
		mockPost: func(url string, contentType string, reader io.Reader) (*http.Response, error) {
			expectedURL := fmt.Sprintf("%s/tx", apiURL)
			if url != expectedURL {
				t.Errorf("unexpected url\nexpected: %s\nactual:   %s", expectedURL, url)
			}

			if contentType != "text/plain" {
				t.Errorf("unexpected type\nexpected: text/plain\nactual:   %s", contentType)
			}

			readerBuffer := new(strings.Builder)
			_, err := io.Copy(readerBuffer, reader)
			if err != nil {
				t.Fatal(err)
			}
			if readerBuffer.String() != transaction {
				t.Errorf("unexpected transaction in body\nexpected: %s\nactual:   %s", transaction, readerBuffer.String())
			}

			return &http.Response{
				StatusCode: 200,
			}, nil
		},
	})
	err := electrs.Broadcast(transaction)
	if err != nil {
		t.Error(err)
	}
}

func TestElectrsConnection_vbyteFee(t *testing.T) {
	apiURL := "example.org/api"
	electrs := NewElectrsConnection(apiURL)
	jsonResponse := `{ "1": 87.882, "2": 87.882, "3": 87.882, "4": 87.882, "5": 81.129, "6": 68.285, "7": 65.182, "8": 63.876, "9": 61.153, "10": 60.172, "11": 57.721, "12": 54.753, "13": 52.879, "14": 46.872, "15": 42.871, "16": 39.989, "17": 35.919, "18": 30.821, "19": 25.888, "20": 21.876, "21": 16.156, "22": 11.222, "23": 10.982, "24": 9.654, "25": 7.883, "144": 1.027, "504": 1.027, "1008": 1.027 }`
	electrs.setClient(mockClient{
		mockGet: func(url string) (*http.Response, error) {
			expectedURL := fmt.Sprintf("%s/fee-estimates", apiURL)
			if url != expectedURL {
				t.Errorf("unexpected url\nexpected: %s\nactual:   %s", expectedURL, url)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse))),
			}, nil
		},
		mockPost: func(url string, contentType string, reader io.Reader) (*http.Response, error) {
			return nil, nil
		},
	})
	expectedFee := int32(7)
	fee, err := electrs.VbyteFee()
	if err != nil {
		t.Fatal(err)
	}
	if fee != expectedFee {
		t.Errorf("unexpected fee\nexpected: %d\nactual:   %d", expectedFee, fee)
	}
}
