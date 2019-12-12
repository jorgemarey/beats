package omega

import (
	"context"
	"fmt"
	"net/http"
)

// Load sends a bulk of logs entries to omega
func (c *Client) Load(ctx context.Context, entries []*LogEntry) error {
	url := fmt.Sprintf("ns/%s/logs", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(ctx, http.MethodPost, url, entries) }
	return c.do(r, nil)
}
