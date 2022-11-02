package dbaasclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	HTTPClient   *retryablehttp.Client
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	Timeout      time.Duration
}

type providerResponse struct {
	Result struct {
		Found bool `json:"found"`
	} `json:"result"`
	Error string `json:"error"`
}

func addProtocol(url string) string {
	if !strings.Contains(url, "https://") {
		if !strings.Contains(url, "http://") {
			return fmt.Sprintf("http://%s", url)
		}
	}
	return url
}

func NewClient(c Client) *Client {
	httpClient := retryablehttp.NewClient()
	// set up the default retries
	httpClient.RetryMax = 5
	if c.RetryMax > 0 {
		httpClient.RetryMax = c.RetryMax
	}
	// set the default retry wait minimum to 1s
	httpClient.RetryWaitMin = time.Duration(1000) * time.Millisecond
	if c.RetryWaitMin > 0 {
		httpClient.RetryWaitMin = c.RetryWaitMin
	}
	// set the default retry wait maximum to 5s
	httpClient.RetryWaitMax = time.Duration(5000) * time.Millisecond
	if c.RetryWaitMax > 0 {
		httpClient.RetryWaitMax = c.RetryWaitMax
	}
	// set the http client timeout to 10s
	httpClient.HTTPClient.Timeout = time.Duration(10000) * time.Millisecond
	if c.Timeout > 0 {
		httpClient.HTTPClient.Timeout = c.Timeout
	}
	// disable the retryablehttp client logger
	httpClient.Logger = nil
	c.HTTPClient = httpClient
	return &c
}

func (c *Client) CheckHealth(dbaasEndpoint string) error {
	// curl --write-out "%{http_code}\n" --silent --output /dev/null "http://dbaas/healthz"
	dbaasEndpoint = addProtocol(dbaasEndpoint)
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/healthz", dbaasEndpoint))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// check the dbaas provider exists, will return true or false without error if it can talk to the dbaas-operator
// will return error if there an issue with the dbaas-operator or the specified endpoint
func (c *Client) CheckProvider(dbaasEndpoint, dbaasType, dbaasEnvironment string) (bool, error) {
	dbaasEndpoint = addProtocol(dbaasEndpoint)
	// curl --silent "http://dbaas/type/env"
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/%s/%s", dbaasEndpoint, dbaasType, dbaasEnvironment))
	if err != nil {
		return false, err
	}
	response := new(providerResponse)
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return false, fmt.Errorf("dbaas operator responded, but response is not a valid JSON payload")
	}
	if response.Error != "" {
		return false, fmt.Errorf(response.Error)
	}
	if response.Result.Found {
		return true, nil
	}
	return false, nil
}

// TestDBaaSHTTPServer is a test server used to test dbaas-responses
func TestDBaaSHTTPServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/mariadb/production", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/mariadb/development", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/mariadb/development2", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":false},"error":"no providers for dbaas environment development2"}`))
	})
	mux.HandleFunc("/postgres/production", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/postgres/development", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/postgres/development2", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":false},"error":"no providers for dbaas environment development2"}`))
	})
	mux.HandleFunc("/mongodb/production", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/mongodb/development", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":true}}`))
	})
	mux.HandleFunc("/mongodb/development2", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"result":{"found":false},"error":"no providers for dbaas environment development2"}`))
	})
	mux.HandleFunc("/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("{}"))
	})
	ts := httptest.NewServer(mux)
	return ts
}
