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
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"

	"github.com/kr/logfmt"
)

type docker_log_parser struct{}

func init() {
	processors.RegisterPlugin("docker_log_parser", NewDockerLogParser)
}

func NewDockerLogParser(c *common.Config) (processors.Processor, error) {
	return &docker_log_parser{}, nil
}

type dockerLogEntry struct {
	Time  time.Time
	Level string
	Msg   string
	Error string
}

func (p *docker_log_parser) Run(event *beat.Event) (*beat.Event, error) {
	msg, err := event.GetValue("message")
	if err != nil {
		return event, nil
	}

	dentry := &dockerLogEntry{}
	if err := logfmt.Unmarshal([]byte(msg.(string)), &dentry); err != nil {
		return event, nil
	}

	event.PutValue("semaas.log.level", getLevelStr(dentry.Level))
	event.PutValue("message", dentry.Msg)
	if dentry.Error != "" {
		event.PutValue("message", fmt.Sprintf("%s (%s)", dentry.Msg, dentry.Error))
	}
	event.Timestamp = dentry.Time

	return event, nil
}

func (p docker_log_parser) String() string {
	return "docker_log_parser="
}

func (e *dockerLogEntry) HandleLogfmt(key, val []byte) error {
	switch k := string(key); k {
	case "time":
		t, err := time.Parse(time.RFC3339Nano, string(val))
		if err != nil {
			return fmt.Errorf("Cant parse time: %s", err)
		}
		e.Time = t
	case "level":
		e.Level = string(val)
	case "msg":
		e.Msg = string(val)
	case "error":
		e.Error = string(val)
	}
	return nil
}
