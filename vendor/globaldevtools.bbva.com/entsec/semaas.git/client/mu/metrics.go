package mu

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Metric

func (c *Client) CreateMetric(metric MetricSpec) (string, error) {
	url := fmt.Sprintf("ns/%s/metrics", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodPost, url, metric) }
	return metric.Locator, c.do(r, &metric)
}

func (c *Client) GetMetric(metricID string) (MetricSpec, error) {
	url := fmt.Sprintf("ns/%s/metrics/%s", c.c.Namespace(), metricID)
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	var result MetricSpec
	return result, c.do(r, &result)
}

func (c *Client) ListMetrics() ([]MetricSpec, error) {
	url := fmt.Sprintf("ns/%s/metrics", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	metrics := struct {
		Metrics []MetricSpec `json:"metrics"`
	}{}
	return metrics.Metrics, c.do(r, &metrics)
}

// MetricSetType

func (c *Client) CreateMetricSetType(metricSetType MetricSetType) (string, error) {
	url := fmt.Sprintf("ns/%s/metric-set-types", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodPost, url, metricSetType) }
	return metricSetType.Locator, c.do(r, &metricSetType)
}

func (c *Client) GetMetricSetType(metricSetTypeID string) (MetricSetType, error) {
	url := fmt.Sprintf("ns/%s/metric-set-types/%s", c.c.Namespace(), metricSetTypeID)
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	var result MetricSetType
	return result, c.do(r, &result)
}

func (c *Client) ListMetricSetTypes() ([]MetricSetType, error) {
	url := fmt.Sprintf("ns/%s/metric-set-types", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	result := struct {
		MetricSetTypes []MetricSetType `json:"metricSetTypes"`
	}{}
	return result.MetricSetTypes, c.do(r, &result)
}

func (c *Client) DeleteMetricSetType(metricSetTypeID string) error {
	url := fmt.Sprintf("ns/%s/metric-set-types/%s", c.c.Namespace(), metricSetTypeID)
	r := func() (*http.Request, error) { return c.request(nil, http.MethodDelete, url, nil) }
	return c.do(r, nil)
}

// MetricSet

func (c *Client) CreateMetricSet(metricSet MetricSet) error {
	url := fmt.Sprintf("ns/%s/metric-sets", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodPost, url, metricSet) }
	return c.do(r, nil)
}

func (c *Client) GetMetricSet(metricSetID string) (MetricSet, error) {
	url := fmt.Sprintf("ns/%s/metric-sets/%s", c.c.Namespace(), metricSetID)
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	var result MetricSet
	return result, c.do(r, &result)
}

func (c *Client) ListMetricSets() ([]MetricSet, error) {
	url := fmt.Sprintf("ns/%s/metric-sets", c.c.Namespace())
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	metricSets := struct {
		MetricSets []MetricSet `json:"metricSets"`
	}{}
	return metricSets.MetricSets, c.do(r, &metricSets)
}

func (c *Client) DeleteMetricSet(metricSetID string) error {
	url := fmt.Sprintf("ns/%s/metric-sets/%s", c.c.Namespace(), metricSetID)
	r := func() (*http.Request, error) { return c.request(nil, http.MethodDelete, url, nil) }
	return c.do(r, nil)
}

type metricsMeasurements struct {
	Metrics []Metrics `json:"metrics"`
}

func (c *Client) AddMeasurements(metricSetID string, metrics []Metrics, opts ...RequestOption) error {
	options := requestOptions{namespace: c.c.Namespace()}
	for _, opt := range opts {
		opt(&options)
	}
	url := fmt.Sprintf("ns/%s/metric-sets/%s:addMeasurements", options.namespace, metricSetID)
	m := metricsMeasurements{Metrics: metrics}
	r := func() (*http.Request, error) { return c.request(nil, http.MethodPost, url, m) }
	return c.do(r, nil)
}

func (c *Client) ListMeasurements(metricSetID string, opts ...QueryOption) ([]Metrics, error) {
	url := c.listURL(metricSetID, "listMeasurements", opts)
	m := metricsMeasurements{}
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	return m.Metrics, c.do(r, &m)
}

type metricsTimeseries struct {
	Timeseries []Timeseries `json:"timeseries"`
}

func (c *Client) ListTimeseries(metricSetID string, opts ...QueryOption) ([]Timeseries, error) {
	url := c.listURL(metricSetID, "listTimeseries", opts)
	m := metricsTimeseries{}
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	return m.Timeseries, c.do(r, &m)
}

type metricsAggregations struct {
	Buckets []Aggregation `json:"buckets"`
}

func (c *Client) ListAggregations(metricSetID string, opts ...QueryOption) ([]Aggregation, error) {
	url := c.listURL(metricSetID, "listAggregations", opts)
	m := metricsAggregations{}
	r := func() (*http.Request, error) { return c.request(nil, http.MethodGet, url, nil) }
	return m.Buckets, c.do(r, &m)
}

func (c *Client) listURL(metricSetID, kind string, opts []QueryOption) string {
	q := &queryOptions{Values: url.Values{}, aggregates: make([]string, 0), operations: make(map[string][]string)}
	for _, opt := range opts {
		opt(q)
	}
	q.prepare()

	url := fmt.Sprintf("ns/%s/metric-sets/%s:%s", c.c.Namespace(), metricSetID, kind)
	if qe := q.Encode(); qe != "" {
		url = fmt.Sprintf("%s?%s", url, qe)
	}
	return url
}

// Query options

// QueryOption configures how we do the query
type QueryOption func(*queryOptions)

type queryOptions struct {
	url.Values
	aggregates []string
	operations map[string][]string
}

func (o *queryOptions) prepare() {
	if o.aggregates != nil && len(o.aggregates) > 0 {
		o.Set("aggregate", strings.Join(o.aggregates, ","))
	}
	if o.operations != nil {
		for k, v := range o.operations {
			o.Add("operation", fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
		}
	}
}

// QueryFrom sets an initial time for the results
func QueryFrom(from time.Time) QueryOption {
	return func(o *queryOptions) {
		o.Set("fromTimestamp", strconv.FormatInt(from.UnixNano(), 10))
	}
}

// QueryTo sets an end time for the results
func QueryTo(to time.Time) QueryOption {
	return func(o *queryOptions) {
		o.Set("toTimestamp", strconv.FormatInt(to.UnixNano(), 10))
	}
}

// QueryAggregating sets aggregation parameters
func QueryAggregating(param string, params ...string) QueryOption {
	return func(o *queryOptions) {
		o.aggregates = append(o.aggregates, param)
		o.aggregates = append(o.aggregates, params...)
	}
}

// QueryOperation sets aggregation parameters
func QueryOperation(op, param string, params ...string) QueryOption {
	return func(o *queryOptions) {
		v, ok := o.operations[op]
		if !ok {
			v = make([]string, 0, len(params)+1)
		}
		v = append(v, param)
		v = append(v, params...)
		o.operations[op] = v
	}
}

// QueryGranularity sets the fetch duration between data
func QueryGranularity(duration time.Duration) QueryOption {
	return func(o *queryOptions) {
		o.Set("granularity", duration.String())
	}
}

// QueryRSQL sets a filter to the search
func QueryRSQL(query string) QueryOption {
	return func(o *queryOptions) {
		o.Set("q", query)
	}
}

// Option sets options for the request
type RequestOption func(*requestOptions)

type requestOptions struct {
	namespace string
}

// WithNamespace sets the namespace where to make the request
func WithNamespace(namespace string) RequestOption {
	return func(o *requestOptions) {
		o.namespace = namespace
	}
}
