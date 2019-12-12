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
	"encoding/json"
	"time"
)

// Span represents a contiguous segment of work in a Trace
type Span struct {
	MrID       string                 `json:"mrId"`
	Name       string                 `json:"name,omitempty"`
	SpanID     string                 `json:"spanId,omitempty"`
	TraceID    string                 `json:"traceId,omitempty"`
	ParentSpan string                 `json:"parentSpan,omitempty"`
	StartDate  time.Time              `json:"-"`
	FinishDate time.Time              `json:"-"`
	Duration   int64                  `json:"duration,omitempty"`
	RecordDate time.Time              `json:"recordDate,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type spanInternal Span // just to avoid a marshall loop

type jsonSpan struct {
	spanInternal
	StartDate  int64 `json:"startDate,omitempty"`
	FinishDate int64 `json:"finishDate,omitempty"`
	RecordDate int64 `json:"recordDate,omitempty"`
}

func (m *Span) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		jsonSpan{
			spanInternal: spanInternal(*m),
			StartDate:    m.StartDate.UnixNano(),
			FinishDate:   m.FinishDate.UnixNano(),
			RecordDate:   m.RecordDate.UnixNano(),
		},
	)
}

func (m *Span) UnmarshalJSON(data []byte) error {
	var v jsonSpan
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Span(v.spanInternal)
	m.StartDate = time.Unix(0, v.StartDate)
	m.FinishDate = time.Unix(0, v.FinishDate)
	m.RecordDate = time.Unix(0, v.RecordDate)
	return nil
}
