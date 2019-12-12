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

type vault_audit_parser struct{}

func init() {
	processors.RegisterPlugin("vault_audit_parser", NewVaultAuditParser)
}

func NewVaultAuditParser(c *common.Config) (processors.Processor, error) {
	return &vault_audit_parser{}, nil
}

func (p *vault_audit_parser) Run(event *beat.Event) (*beat.Event, error) {
	msg, err := event.GetValue("message")
	if err != nil {
		return event, nil
	}

	data := map[string]interface{}{}
	if err := json.Unmarshal([]byte(msg.(string)), &data); err != nil {
		return event, nil
	}

	if it, ok := data["time"]; ok {
		if tStr, ok := it.(string); ok {
			t, err := time.Parse(time.RFC3339, tStr)
			if err != nil {
				return event, nil
			}
			event.Timestamp = t
		}
	}
	parseMap("semaas.properties", "V", data, event)

	return event, nil
}

func (p vault_audit_parser) String() string {
	return "vault_audit_parser="
}
