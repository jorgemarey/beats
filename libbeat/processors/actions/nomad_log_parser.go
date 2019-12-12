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
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"
)

type nomad_log_parser struct{}

func init() {
	processors.RegisterPlugin("nomad_log_parser", NewNomadLogParser)
}

func NewNomadLogParser(c *common.Config) (processors.Processor, error) {
	return &nomad_log_parser{}, nil
}

func (p *nomad_log_parser) Run(event *beat.Event) (*beat.Event, error) {
	msg, err := event.GetValue("message")
	if err != nil {
		return event, nil
	}

	data := map[string]interface{}{}
	if err := json.Unmarshal([]byte(msg.(string)), &data); err != nil {
		return event, nil
	}

	for k, v := range data {
		switch k {
		case "@message":
			event.PutValue("message", v.(string))
		case "@level":
			event.PutValue("semaas.log.level", getLevelStr(v.(string)))
		case "@timestamp":
			t, err := time.Parse("2006-01-02T15:04:05.000000Z07:00", v.(string))
			if err != nil {
				return event, nil
			}
			event.Timestamp = t
		case "@module":
			event.PutValue("semaas.properties.module", v.(string))
		// case "@caller": // This has file and line information in each log line (won't come in production logs)
		default:
			parseItem("semaas.properties", k, v, event)
		}
	}

	return event, nil
}

func (p nomad_log_parser) String() string {
	return "nomad_log_parser="
}
