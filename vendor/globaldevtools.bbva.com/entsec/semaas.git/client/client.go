// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	envNamespace = "SEMAAS_NAMESPACE"
	envAPIKey    = "SEMAAS_API_KEY"
)

// Client represent a client to the semaas APIs
type Client struct {
	options    clientOptions
	httpClient *http.Client
}

// New creates a default client with the provided options
func New(opts ...Option) (*Client, error) {
	c := &Client{httpClient: http.DefaultClient}
	c.setDefaultOptions()
	for _, opt := range opts {
		opt(&c.options)
	}
	if c.options.namespace == "" {
		return nil, fmt.Errorf("Namespace value must be provided")
	}
	if fn := c.options.authFn; fn != nil {
		auth, err := fn()
		if err != nil {
			return nil, fmt.Errorf("Can't perform authentication: %s", err)
		}
		c.options.authorization = auth
	}
	c.httpClient.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &c.options.tlsConfig,
	}
	return c, nil
}

func (c *Client) setDefaultOptions() {
	defaultOpts := []Option{
		WithAPIKey(os.Getenv(envAPIKey)),
		WithNamespace(os.Getenv(envNamespace)),
	}
	for _, opt := range defaultOpts {
		opt(&c.options)
	}
}

// Namespace returns the namespace name set on the client
func (c *Client) Namespace() string {
	return c.options.namespace
}

func (c *Client) requestWithReader(ctx context.Context, method, url string, reader io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	if c.options.apiKey != "" {
		req.Header.Add("Api-Key", c.options.apiKey)
	}
	if c.options.authorization != "" {
		req.Header.Add("Authorization", c.options.authorization)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	return req, nil
}

// Request builds a new http.Request out of the provided parameters
func (c *Client) Request(ctx context.Context, method, path string, data interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)
	if data != nil {
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return nil, fmt.Errorf("Error encoding data: %s", err)
		}
	}
	return c.requestWithReader(ctx, method, fmt.Sprintf("%s%s", c.options.url, path), buf)
}

// Requester returns a request to be executed or an error
type Requester func() (*http.Request, error)

// ResponseChecker checks whether the response if OK or and error should be thrown
type ResponseChecker interface {
	Check(*http.Response) error
}

// Do executes the Requester and process the response returning the data on the argument if everything is ok
func (c *Client) Do(requester Requester, checker ResponseChecker, data interface{}) error {
	req, err := requester()
	if err != nil {
		return err
	}
	if c.options.debug {
		breq, _ := httputil.DumpRequest(req, true)
		log.Print(string(breq))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request %s: %s", req.URL, err)
	}
	if c.options.debug {
		bresp, _ := httputil.DumpResponse(resp, true)
		log.Print(string(bresp))
	}
	defer resp.Body.Close()
	if checker != nil {
		if err = checker.Check(resp); err != nil {
			return err
		}
	} else {
		if resp.StatusCode >= 400 {
			bb, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Status code was %d, couldn't read body: %s", resp.StatusCode, err)
			}
			return fmt.Errorf("Status code was %d: %s", resp.StatusCode, string(bb))
		}
	}
	if data != nil && resp.StatusCode != http.StatusNoContent {
		err = json.NewDecoder(resp.Body).Decode(data)
	}
	return err
}
