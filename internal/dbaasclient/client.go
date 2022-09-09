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

var httpClient = retryablehttp.NewClient()

var retryWaitMin = 1000
var retryWaitMax = 5000

type Client struct {
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	Timeout      time.Duration
}

func NewClient(c Client) *Client {
	if c.RetryMax == 0 {
		c.RetryMax = 5
	}
	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = time.Duration(1) * time.Second
	}
	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = time.Duration(5) * time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = time.Duration(1) * time.Second
	}
	return &c
}

func addProtocol(url string) string {
	if !strings.Contains(url, "https://") {
		if !strings.Contains(url, "http://") {
			return fmt.Sprintf("http://%s", url)
		}
	}
	return url
}

func (c *Client) CheckDBaaSHealth(dbaasEndpoint string) error {
	// curl --write-out "%{http_code}\n" --silent --output /dev/null "http://dbaas/healthz"
	dbaasEndpoint = addProtocol(dbaasEndpoint)
	httpClient.RetryMax = c.RetryMax
	httpClient.RetryWaitMin = c.RetryWaitMin
	httpClient.RetryWaitMax = c.RetryWaitMin
	httpClient.HTTPClient.Timeout = c.Timeout
	httpClient.Logger = nil
	resp, err := httpClient.Get(fmt.Sprintf("%s/healthz", dbaasEndpoint))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type providerResponse struct {
	Result struct {
		Found bool `json:"found"`
	} `json:"result"`
	Error string `json:"error"`
}

// check the dbaas provider exists, will return true or false without error if it can talk to the dbaas-operator
// will return error if there an issue with the dbaas-operator or the specified endpoint
func (c *Client) CheckDBaaSProvider(dbaasEndpoint, dbaasType, dbaasEnvironment string) (bool, error) {
	dbaasEndpoint = addProtocol(dbaasEndpoint)
	// curl --silent "http://dbaas/type/env"
	httpClient.RetryMax = c.RetryMax
	httpClient.RetryWaitMin = c.RetryWaitMin
	httpClient.RetryWaitMax = c.RetryWaitMin
	httpClient.HTTPClient.Timeout = c.Timeout
	httpClient.Logger = nil
	resp, err := httpClient.Get(fmt.Sprintf("%s/%s/%s", dbaasEndpoint, dbaasType, dbaasEnvironment))
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
