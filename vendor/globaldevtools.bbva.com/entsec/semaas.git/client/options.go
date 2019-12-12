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

package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// Token authentication
type Token struct {
	Accesstoken string `json:"access_token"`
}

// Api logic
type clientOptions struct {
	url           string
	namespace     string
	apiKey        string
	authorization string
	authFn        func() (string, error)
	debug         bool
	tlsConfig     tls.Config
}

// Option sets options over the client
type Option func(*clientOptions)

// WithDebug sets the Client to debug the requests and responses
func WithDebug() Option {
	return func(o *clientOptions) {
		o.debug = true
	}
}

// WithURL sets the URL where to make the load requests
func WithURL(URL string) Option {
	return func(o *clientOptions) {
		o.url = URL
	}
}

// WithNamespace sets the namespace making the requests
func WithNamespace(namespace string) Option {
	return func(o *clientOptions) {
		o.namespace = namespace
	}
}

// WithAPIKey sets the Api-Key to make authenticated requests
func WithAPIKey(APIKey string) Option {
	return func(o *clientOptions) {
		o.apiKey = APIKey
	}
}

// WithSkipVerify skips TLS verification when comunicating with the server
func WithSkipVerify() Option {
	return func(o *clientOptions) {
		o.tlsConfig.InsecureSkipVerify = true
	}
}

// WithServerCerts Allows you to set multiple certs instead of just one. Deprecates WithCert
func WithServerCerts(cacerts ...*x509.Certificate) Option {
	return func(o *clientOptions) {
		if o.tlsConfig.RootCAs == nil {
			if pool, err := x509.SystemCertPool(); err != nil {
				o.tlsConfig.RootCAs = x509.NewCertPool()
			} else {
				o.tlsConfig.RootCAs = pool
			}
		}
		for _, cert := range cacerts {
			o.tlsConfig.RootCAs.AddCert(cert)
		}
	}
}

// WithClientCert set a client certificate to perform authentication.
// Can be read of a file using tls.LoadX509KeyPair()
func WithClientCert(cert tls.Certificate) Option {
	return func(o *clientOptions) {
		o.tlsConfig.Certificates = []tls.Certificate{cert}
	}
}

// WithIdpCredentials set a Idp-credentials Header to authenticate with.
func WithIdpCredentials(user, pass, mrURL string) Option {
	return func(o *clientOptions) {
		fn := func() (string, error) {
			c := fmt.Sprintf("%s:%s", user, pass)
			basic := fmt.Sprintf("%s %s", "Basic", base64.StdEncoding.EncodeToString([]byte(c)))
			authurl := fmt.Sprintf("%s/auth/token?grant_type=client_credentials", mrURL)
			client := &http.Client{}
			r, err := http.NewRequest("POST", authurl, nil)
			if err != nil {
				return "", err
			}
			r.Header.Add("Authorization", basic)
			resp, err := client.Do(r)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			var msg Token
			if err = json.NewDecoder(resp.Body).Decode(&msg); err != nil {
				return "", err
			}
			return fmt.Sprintf("%s %s", "JWT", msg.Accesstoken), nil
		}
		o.authFn = fn
	}
}

// TODO: add withTransport for changing some roundtrip parameters
