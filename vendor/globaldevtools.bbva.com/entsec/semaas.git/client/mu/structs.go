package mu

import (
	"encoding/json"
	"time"
)

type DataType string

// Data types available
const (
	DataTypeFloat   DataType = "float"
	DataTypeInteger DataType = "integer"
)

type Unit string

// Unit values availables
const (
	UnitPercent  Unit = "PERCENT"
	UnitCount    Unit = "COUNT"
	UnitSeconds  Unit = "SECONDS"
	UnitMicrosec Unit = "MICROSECONDS"
	UnitMillisec Unit = "MILLISECONDS"
	UnitBytes    Unit = "BYTES"
	UnitKBytes   Unit = "KILOBYTES"
	UnitMBytes   Unit = "MEGABYTES"
	UnitGBytes   Unit = "GIGABYTES"
	UnitTBytes   Unit = "TERABYTES"
	UnitBits     Unit = "BITS"
	UnitKBits    Unit = "KILOBITS"
	UnitMBits    Unit = "MEGABITS"
	UnitGBits    Unit = "GIGABITS"
	UnitTBits    Unit = "TERABITS"
	UnitBSec     Unit = "BYTES/SECOND"
	UnitKBSec    Unit = "KILOBYTES/SECOND"
	UnitMBSec    Unit = "MEGABYTES/SECOND"
	UnitGBSec    Unit = "GIGABYTES/SECOND"
	UnitTBSec    Unit = "TERABYTES/SECOND"
	UnitbSec     Unit = "BITS/SECOND"
	UnitKbSec    Unit = "KILOBITS/SECOND"
	UnitMbSec    Unit = "MEGABITS/SECOND"
	UnitGbSec    Unit = "GIGABITS/SECOND"
	UnitTbSec    Unit = "TERABITS/SECOND"
	UnitNone     Unit = "NONE"
)

type MetricSpec struct {
	ID          string   `json:"_id"`
	Locator     string   `json:"_locator,omitempty"`
	DataType    DataType `json:"dataType"`
	DataUnit    Unit     `json:"dataUnit"`
	Description string   `json:"description"`
}

type MetricSetType struct {
	ID         string            `json:"_id"`
	Locator    string            `json:"_locator,omitempty"`
	MetricSpec map[string]string `json:"metricsSpec"` //map<string,string(metricSpec._locator)>
}

type MetricSet struct { // TODO: The field monitoredResource when retrieved is an object
	ID                string      `json:"_id"`
	MetricSetType     string      `json:"metricSetType"`     // string(metricSetType._locator)
	MonitoredResource interface{} `json:"monitoredResource"` // string(monitoredResource._locator)
}

type MetricValue interface{} // This can only be a number (int or float)

type Metrics struct {
	Timestamp  time.Time              `json:"-"`
	Values     map[string]MetricValue `json:"values"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type Timeseries struct {
	Metrics     []Metrics              `json:"metrics"`
	Aggregation map[string]interface{} `json:"aggregation"`
}

type Aggregation struct {
	Values map[string]MetricValue `json:"values"`
	Bucket map[string]interface{} `json:"bucket"`
}

type metricInternal Metrics // just to avoid a marshall loop

type jsonMetrics struct {
	metricInternal
	Timestamp int64 `json:"timestamp"`
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		jsonMetrics{
			metricInternal: metricInternal(*m),
			Timestamp:      m.Timestamp.UnixNano(),
		},
	)
}

func (m *Metrics) UnmarshalJSON(data []byte) error {
	var v jsonMetrics
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Metrics(v.metricInternal)
	m.Timestamp = time.Unix(0, v.Timestamp)
	return nil
}
