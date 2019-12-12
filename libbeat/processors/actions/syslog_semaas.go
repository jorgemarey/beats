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
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"
)

const (
	severityField = "event.severity"
)

type syslog_semaas struct{}

func init() {
	processors.RegisterPlugin("syslog_semaas", NewSyslogSemaas)
}

func NewSyslogSemaas(c *common.Config) (processors.Processor, error) {
	return &syslog_semaas{}, nil
}

func (p *syslog_semaas) Run(event *beat.Event) (*beat.Event, error) {
	if fieldValue, err := event.GetValue(severityField); err == nil {
		if value, ok := fieldValue.(int); ok {
			event.PutValue("semaas.log.level", getLevel(value))
		}
	}

	if fieldValue, err := event.GetValue("hostname"); err == nil {
		event.PutValue("semaas.properties.hostname", fieldValue)
		event.PutValue("semaas.properties.source", fieldValue) // Just for backwards compatibility
	}

	if fieldValue, err := event.GetValue("process.program"); err == nil {
		event.PutValue("semaas.properties.service", fieldValue)
	}

	if fieldValue, err := event.GetValue("syslog.facility_label"); err == nil {
		event.PutValue("semaas.properties.facility", fieldValue)
	}

	if fieldValue, err := event.GetValue("syslog.facility"); err == nil {
		event.PutValue("semaas.properties.facility_id", fieldValue)
	}

	if fieldValue, err := event.GetValue("process.pid"); err == nil {
		event.PutValue("semaas.properties.pid", fieldValue)
	}

	return event, nil
}

func (p syslog_semaas) String() string {
	return "syslog_semaas="
}