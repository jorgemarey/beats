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

// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package semaas

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs"
)

func init() {
	outputs.RegisterType("semaas", makeSemaas)
}

func makeSemaas(_ outputs.IndexManager, beat beat.Info, observer outputs.Observer, cfg *common.Config) (outputs.Group, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	if config.KeyFile == "" || config.CertFile == "" {
		outputs.Fail(errors.New("both key_file and cert_file must be provided"))
	}

	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		outputs.Fail(fmt.Errorf("error loading certificate: %s", err))
	}

	c, err := newClient(cert, config.Namespace, config.MrID, config.NamespaceField, config.MrIDField, config.OmegaURL, config.RhoURL, config.AdditionalPropertyFields, config.Timeout)
	if err != nil {
		outputs.Fail(fmt.Errorf("error creating client: %s", err))
	}

	return outputs.Success(config.BulkMaxSize, config.MaxRetries, c)
}
