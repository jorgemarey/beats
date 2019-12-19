package mu

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"globaldevtools.bbva.com/entsec/semaas.git/client"
)

const (
	envURL = "MU_URL"
)

// EnvOptions returns some default options readed from environment variables
func EnvOptions() []client.Option {
	return []client.Option{
		client.WithURL(os.Getenv(envURL)),
	}
}

// Client is a Omega client used to perform Omega API requests
type Client struct {
	c *client.Client
}

// New creates a Mu client with provided options
func New(c *client.Client) (*Client, error) {
	if c == nil {
		return nil, fmt.Errorf("A client should be provided")
	}
	cli := &Client{c: c}
	return cli, nil
}

func (c *Client) request(ctx context.Context, method, path string, data interface{}) (*http.Request, error) {
	return c.c.Request(ctx, method, fmt.Sprintf("/v0/%s", path), data)
}

func (c *Client) do(requester client.Requester, data interface{}) error {
	return c.c.Do(requester, nil, data)
}
