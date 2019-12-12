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

package semaas

import "time"

type semaasConfig struct {
	CertFile                 string        `config:"cert_file"`
	KeyFile                  string        `config:"key_file"`
	OmegaURL                 string        `config:"omega_url"`
	RhoURL                   string        `config:"rho_url"`
	MrID                     string        `config:"mrID"`
	MrIDField                string        `config:"mrID_field"`
	Namespace                string        `config:"namespace"`
	NamespaceField           string        `config:"namespace_field"`
	AdditionalPropertyFields []string      `config:"additional_property_fields"`
	BulkMaxSize              int           `config:"bulk_max_size"`
	MaxRetries               int           `config:"max_retries"`
	Timeout                  time.Duration `config:"timeout"`
	// Backoff     backoff               `config:"backoff"`
}

func defaultConfig() *semaasConfig {
	return &semaasConfig{
		Timeout:     10 * time.Second,
		BulkMaxSize: 2048,
		MaxRetries:  3,
	}
}

func (c *semaasConfig) Validate() error {
	return nil
}
