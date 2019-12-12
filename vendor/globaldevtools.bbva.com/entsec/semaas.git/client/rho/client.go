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

package rho

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"globaldevtools.bbva.com/entsec/semaas.git/client"
	"globaldevtools.bbva.com/entsec/semaas.git/client/omega"
)

const (
	envURL = "RHO_URL"
)

// EnvOptions returns some default options readed from environment variables
func EnvOptions() []client.Option {
	return []client.Option{
		client.WithURL(os.Getenv(envURL)),
	}
}

// Client is a Rho client used to perform Rho API requests
type Client struct {
	c *client.Client
}

// New creates an Rho client with provided options
func New(c *client.Client) (*Client, error) {
	if c == nil {
		return nil, fmt.Errorf("A client should be provided")
	}
	cli := &Client{c: c}

	return cli, nil
}

func (c *Client) request(ctx context.Context, method, path string, data interface{}) (*http.Request, error) {
	return c.c.Request(ctx, method, fmt.Sprintf("/v1/%s", path), data)
}

func (c *Client) requestCompletePath(ctx context.Context, method, path string, data interface{}) (*http.Request, error) {
	return c.c.Request(ctx, method, path, data)
}

func (c *Client) do(requester client.Requester, data interface{}) error {
	return c.c.Do(requester, &rhoErrorChecker{}, data)
}

type rhoErrorChecker struct{}

func (c *rhoErrorChecker) Check(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		omegaErr := &omega.LoadError{}
		bb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Status code was %d, couldn't read body: %s", resp.StatusCode, err)
		}
		if err := json.Unmarshal(bb, omegaErr); err != nil {
			return fmt.Errorf("Can't get coded error: %s. Error status '%d': %s", err, resp.StatusCode, string(bb))
		}
		if omegaErr.Code == 0 {
			return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
		if len(omegaErr.InvalidEntities) == 0 {
			return &omegaErr.Err
		}
		return omegaErr
	}
	return nil
}
