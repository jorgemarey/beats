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

package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"

	"globaldevtools.bbva.com/entsec/semaas.git/client/mu"
	"globaldevtools.bbva.com/entsec/semaas.git/client/omega"
	"globaldevtools.bbva.com/entsec/semaas.git/client/rho"
)

const (
	messageField = "message"
)

type decode_semaas struct {
	logLevel omega.LogLevel
}

func init() {
	processors.RegisterPlugin("decode_semaas", NewDecodeSemaas)
	// configChecked(NewExtractField,
	// 	requireFields("field", "separator", "index", "target"),
	// 	allowedFields("field", "separator", "index", "target", "when")))
}

func NewDecodeSemaas(c *common.Config) (processors.Processor, error) {
	return &decode_semaas{
		logLevel: omega.LogLevelInfo,
	}, nil
}

func (p *decode_semaas) Run(event *beat.Event) (*beat.Event, error) {
	fieldValue, err := event.GetValue(messageField)
	if err != nil {
		return event, fmt.Errorf("error getting field '%s' from event", messageField)
	}

	value, ok := fieldValue.(string)
	if !ok {
		return event, fmt.Errorf("could not get a string from field '%s'", messageField)
	}

	return p.parse(event, value)
}

func (p *decode_semaas) parse(event *beat.Event, message string) (*beat.Event, error) {
	switch {
	case strings.HasPrefix(message, "V2|"):
		parts := strings.Split(message, "|")
		if len(parts) < 3 {
			return event, fmt.Errorf("message isn't in a correct semaas format: does have at least 3 parts")
		}
		kind := parts[1]
		data := strings.Join(parts[2:], "|")

		switch {
		case strings.HasPrefix(kind, "LOG"): // it is a log
			return p.parseLogEntry(event, data, kind)
		case strings.HasPrefix(kind, "SPAN"): // it is a span
			return p.parseSpan(event, data)
		default:
			return event, fmt.Errorf("message isn't in a correct semaas V2 format: kind '%s' should be either LOG or SPAN", kind)
		}
	case strings.HasPrefix(message, "V1|"):
		parts := strings.Split(message, "|")
		if len(parts) < 3 {
			return event, fmt.Errorf("message isn't in a correct semaas format: does have at least 3 parts")
		}
		kind := parts[1]
		data := strings.Join(parts[2:], "|")

		switch {
		case strings.HasPrefix(kind, "METRIC"): // it is a metric
			return p.parseMetricV1(event, data, kind)
		default:
			return event, fmt.Errorf("message isn't in a correct semaas V1 format: kind '%s' should METRIC", kind)
		}
	default:
		// fmt.Errorf("message isn't in a correct semaas format: does not start with version")
		return event, nil
	}
}

func (p *decode_semaas) parseLogEntry(event *beat.Event, data, kind string) (*beat.Event, error) {
	var entry omega.LogEntry
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&entry); err != nil {
		return event, fmt.Errorf("error decoding log entry: %s", err)
	}
	// set the level
	if entry.Level == "" {
		levelParts := strings.Split(kind, ".")
		entry.Level = p.logLevel
		if len(levelParts) > 1 {
			entry.Level = omega.LogLevel(levelParts[1])
		}
	}

	event.PutValue("semaas.kind", "log")

	event.PutValue("message", entry.Message)
	event.PutValue("@timestamp", entry.CreationDate)

	if entry.MrID != "" {
		event.PutValue("semaas.mrId", entry.MrID)
	}
	if entry.SpanID != "" {
		event.PutValue("semaas.spanId", entry.SpanID)
	}
	if entry.TraceID != "" {
		event.PutValue("semaas.traceId", entry.TraceID)
	}
	if entry.Properties != nil && len(entry.Properties) > 0 {
		ok := false
		if propValue, err := event.GetValue("semaas.properties"); err == nil {
			if properties, ok := propValue.(common.MapStr); ok {
				for k, v := range entry.Properties {
					properties[k] = v
				}
				event.PutValue("semaas.properties", properties)
				ok = true
			}
		}
		if !ok {
			event.PutValue("semaas.properties", entry.Properties)
		}
	}
	event.PutValue("semaas.log.level", entry.Level)

	return event, nil
}

func (p *decode_semaas) parseSpan(event *beat.Event, data string) (*beat.Event, error) {
	var span rho.Span
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&span); err != nil {
		return event, fmt.Errorf("error decoding span: %s", err)
	}

	event.PutValue("semaas.kind", "span")

	event.PutValue("message", "-")

	event.PutValue("semaas.mrId", span.MrID)
	event.PutValue("semaas.spanId", span.SpanID)
	event.PutValue("semaas.traceId", span.TraceID)
	if span.Properties != nil && len(span.Properties) > 0 {
		ok := false
		if propValue, err := event.GetValue("semaas.properties"); err == nil {
			if properties, ok := propValue.(common.MapStr); ok {
				for k, v := range span.Properties {
					properties[k] = v
				}
				event.PutValue("semaas.properties", properties)
				ok = true
			}
		}
		if !ok {
			event.PutValue("semaas.properties", span.Properties)
		}
	}

	event.PutValue("semaas.span.name", span.Name)
	event.PutValue("semaas.span.parentSpan", span.ParentSpan)
	event.PutValue("semaas.span.duration", span.Duration)
	event.PutValue("semaas.span.startDate", span.StartDate)
	event.PutValue("semaas.span.finishDate", span.FinishDate)

	return event, nil
}

func (p *decode_semaas) parseMetricV1(event *beat.Event, data, kind string) (*beat.Event, error) {
	var metric Metrics
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&metric); err != nil {
		return event, fmt.Errorf("error decoding metric V1: %s", err)
	}

	semaasKind := "metricv1"
	if kind == "METRIC.CORE" {
		semaasKind = "metriccorev1"
	}
	event.PutValue("semaas.kind", semaasKind)

	if metric.Properties != nil && len(metric.Properties) > 0 {
		ok := false
		if propValue, err := event.GetValue("semaas.properties"); err == nil {
			if properties, ok := propValue.(common.MapStr); ok {
				for k, v := range metric.Properties {
					properties[k] = v
				}
				event.PutValue("semaas.properties", properties)
				ok = true
			}
		}
		if !ok {
			event.PutValue("semaas.properties", metric.Properties)
		}
	}

	event.PutValue("semaas.metric.values", metric.Values)
	event.PutValue("semaas.metric.metricSetId", metric.MetricSetID)

	return event, nil
}

func (p decode_semaas) String() string {
	return "decode_semaas="
}

type Metrics struct {
	Timestamp   time.Time                 `json:"-"`
	MetricSetID string                    `json:"metricSetId"`
	Values      map[string]mu.MetricValue `json:"metrics"`
	Properties  map[string]interface{}    `json:"properties,omitempty"`
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
