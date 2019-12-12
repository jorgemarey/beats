package omega

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"globaldevtools.bbva.com/entsec/semaas.git/client"
)

const (
	envURL = "OMEGA_URL"
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

// New creates an Omenga client with provided options
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
	return c.c.Do(requester, &omegaErrorChecker{}, data)
}

type omegaErrorChecker struct{}

func (c *omegaErrorChecker) Check(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		omegaErr := &LoadError{}
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

// Err represents an error ocurring when calling the Omega API
type Err struct {
	Code    int
	Status  int
	Message string
}

func (e *Err) Error() string {
	return fmt.Sprintf("Error: %s (Code: %d)", e.Message, e.Code)
}

// InvalidEntity contains errors ocurring in the entity of the provided position
type InvalidEntity struct {
	Position int
	Errors   []string // Think this is DEPRECATED
	Error    []string
}

// LoadError represents a specific error that has problems on some entities
type LoadError struct {
	Err
	InvalidEntities []InvalidEntity
}

func (e *LoadError) Error() string {
	ie := make([]string, 0)
	for _, v := range e.InvalidEntities {
		errors := strings.Join(v.Error, " and ")
		if v.Errors != nil && len(v.Errors) > 0 {
			errors = strings.Join(v.Errors, " and ")
		}
		ie = append(ie, fmt.Sprintf("<Item in position %d had the following errors: %s>", v.Position, errors))
	}

	return fmt.Sprintf("Error: %s (Code: %d) Invalid Entities: %s", e.Message, e.Code, strings.Join(ie, ","))
}
