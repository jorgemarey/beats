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

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"globaldevtools.bbva.com/entsec/semaas.git/client"
	"globaldevtools.bbva.com/entsec/semaas.git/client/mu"
	"globaldevtools.bbva.com/entsec/semaas.git/client/omega"
	"globaldevtools.bbva.com/entsec/semaas.git/client/rho"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/publisher"

	"google.golang.org/api/support/bundler"
)

const (
	// DefaultDelayThreshold is the default value for the DelayThreshold Option.
	DefaultDelayThreshold = 10 * time.Second

	// DefaultEntryCountThreshold is the default value for the EntryCountThreshold Option.
	DefaultEntryCountThreshold = 1000

	// DefaultEntryByteThreshold is the default value for the EntryByteThreshold Option.
	DefaultEntryByteThreshold = 1 << 19 // 512KiB

	// DefaultBufferedByteLimit is the default value for the BufferedByteLimit Option.
	DefaultBufferedByteLimit = 1 << 21 // 2MiB
)

const (
	kindLogs  string = "logs"
	kindSpans string = "spans"
)

type semmasClient struct {
	observer outputs.Observer
	timeout  time.Duration
	log 	 *logp.Logger

	omegaURL string
	rhoURL   string
	muURL    string
	muClient *mu.Client

	cert                     tls.Certificate
	defaultMrID              string
	namespace                string
	namespaceField           string
	mrIDField                string
	additionalPropertyFields []string

	semaasBundlers map[string]*bundler.Bundler
	l              sync.Mutex
}

func newClient(cert tls.Certificate, namespace, mrID, namespaceField, mrIDField, omegaURL, rhoURL, muURL string, apf []string, timeout time.Duration) (*semmasClient, error) {
	// opts := omega.EnvOptions()
	// opts = append(opts,
	// 	client.WithClientCert(cert),
	// 	client.WithNamespace(namespace),
	// 	client.WithSkipVerify(),
	// )
	// sc, err := client.New(opts...)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating semaas client: %s", err)
	// }
	// oc, err := omega.New(sc)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating omega client: %s", err)
	// }

	// opts = rho.EnvOptions()
	// opts = append(opts,
	// 	client.WithClientCert(cert),
	// 	client.WithNamespace(namespace),
	// 	client.WithSkipVerify(),
	// )
	// sc, err = client.New(opts...)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating semaas client: %s", err)
	// }
	// rc, err := rho.New(sc)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating rho client: %s", err)
	// }
	mc, err := newMuClient(muURL, cert, timeout)
	if err != nil {
		return nil, err
	}

	return &semmasClient{
		timeout:                  timeout,
		log: 					  logp.NewLogger("semaas"),
		omegaURL:                 omegaURL,
		rhoURL:                   rhoURL,
		muURL:                    muURL,
		muClient:                 mc,
		cert:                     cert,
		defaultMrID:              mrID,
		mrIDField:                mrIDField,
		namespace:                namespace,
		namespaceField:           namespaceField,
		additionalPropertyFields: apf,
		semaasBundlers:           make(map[string]*bundler.Bundler),
	}, nil
}

func (c *semmasClient) Connect() error {
	return nil
}

func (c *semmasClient) Close() error {
	c.l.Lock()
	defer c.l.Unlock()
	wg := sync.WaitGroup{}
	wg.Add(len(c.semaasBundlers))
	for _, v := range c.semaasBundlers {
		go func(v *bundler.Bundler) {
			v.Flush()
			wg.Done()
		}(v)
	}
	wg.Wait()
	return nil
}

func (c *semmasClient) Publish(batch publisher.Batch) error {
	namespace := c.namespace
	mrID := c.defaultMrID

	for _, pev := range batch.Events() {
		event := &pev.Content

		if namespace == "" {
			ns, err := event.GetValue(c.namespaceField)
			if err != nil || ns == "" {
				continue
			}
			namespace = ns.(string)
		}

		if mrID == "" {
			mr, err := event.GetValue(c.mrIDField)
			if err == nil { // Even if this fails we continue
				mrID = mr.(string)
			}
		}

		kind, err := event.GetValue("semaas.kind")
		if err != nil {
			kind = "log"
		}
		switch kind.(string) {
		case "span":
			err = c.processSpan(event, namespace, mrID)
		case "metricv1", "metriccorev1":
			err = c.processMetricV1(event, namespace, mrID, kind.(string))
		case "log":
			err = c.processLog(event, namespace, mrID)
		}
	}

	batch.ACK()
	return nil
}

func (c *semmasClient) String() string {
	return "semmas(" + c.namespace + ")"
}

func (c *semmasClient) processLog(event *beat.Event, namespace, fallbackMrID string) error {
	// parse log
	log := &omega.LogEntry{CreationDate: event.Timestamp}
	if msg, err := event.GetValue("message"); err == nil {
		log.Message = msg.(string)
	}
	// Do not send empty messages
	if log.Message == "" {
		return nil
	}

	if mrID, err := event.GetValue("semaas.mrId"); err == nil {
		log.MrID = mrID.(string)
	}
	if spanID, err := event.GetValue("semaas.spanId"); err == nil {
		log.SpanID = spanID.(string)
	}
	if traceID, err := event.GetValue("semaas.traceId"); err == nil {
		log.TraceID = traceID.(string)
	}
	if level, err := event.GetValue("semaas.log.level"); err == nil {
		log.Level = level.(omega.LogLevel)
	}

	if properties, err := event.GetValue("semaas.properties"); err == nil {
		var ok bool
		log.Properties, ok = properties.(map[string]interface{})
		if !ok {
			props := properties.(common.MapStr)
			log.Properties = props
		}
	}
	if log.Properties == nil {
		log.Properties = make(map[string]interface{})
	}
	for _, field := range c.additionalPropertyFields {
		if properties, err := event.GetValue(field); err == nil {
			for k, v := range properties.(common.MapStr) {
				log.Properties[k] = v
			}
		}
	}

	b := c.getBundler(namespace, kindLogs)
	if b == nil {
		return fmt.Errorf("Unable to get bundler for: %s/%s", namespace, kindLogs)
	}

	if log.MrID == "" {
		log.MrID = fallbackMrID
	}
	if log.Level == "" {
		log.Level = omega.LogLevelInfo
	}

	// TODO: fix len
	if err := b.AddWait(context.Background(), log, len(log.Message)); err != nil {
		return fmt.Errorf("error adding log to bundler: %s", err)
	}
	return nil
}

func (c *semmasClient) processSpan(event *beat.Event, namespace, fallbackMrID string) error {
	// parse span
	span := &rho.Span{}
	if mrID, err := event.GetValue("semaas.mrId"); err == nil {
		span.MrID = mrID.(string)
	}
	if spanID, err := event.GetValue("semaas.spanId"); err == nil {
		span.SpanID = spanID.(string)
	}
	if traceID, err := event.GetValue("semaas.traceId"); err == nil {
		span.TraceID = traceID.(string)
	}

	if name, err := event.GetValue("semaas.span.name"); err == nil {
		span.Name = name.(string)
	}
	if parent, err := event.GetValue("semaas.span.parentSpan"); err == nil {
		span.ParentSpan = parent.(string)
	}

	if duration, err := event.GetValue("semaas.span.duration"); err == nil {
		span.Duration = duration.(int64)
	}
	if startDate, err := event.GetValue("semaas.span.startDate"); err == nil {
		span.StartDate = startDate.(time.Time)
	}
	if finishDate, err := event.GetValue("semaas.span.finishDate"); err == nil {
		span.FinishDate = finishDate.(time.Time)
	}

	if properties, err := event.GetValue("semaas.properties"); err == nil {
		var ok bool
		span.Properties, ok = properties.(map[string]interface{})
		if !ok {
			props := properties.(common.MapStr)
			span.Properties = props
		}
	}
	if span.Properties == nil {
		span.Properties = make(map[string]interface{})
	}
	for _, field := range c.additionalPropertyFields {
		if properties, err := event.GetValue(field); err == nil {
			for k, v := range properties.(common.MapStr) {
				span.Properties[k] = v
			}
		}
	}

	b := c.getBundler(namespace, kindSpans)
	if b == nil {
		return fmt.Errorf("Unable to get bundler for: %s/%s", namespace, kindSpans)
	}

	if span.MrID == "" {
		span.MrID = fallbackMrID
	}

	// TODO: fix len
	if err := b.AddWait(context.Background(), span, len(span.Name)); err != nil {
		return fmt.Errorf("error adding span to bundler: %s", err)
	}
	return nil
}

func (c *semmasClient) getBundler(namespace, kind string) *bundler.Bundler {
	key := namespace + kind
	c.l.Lock()
	defer c.l.Unlock()
	b, ok := c.semaasBundlers[key]
	if ok {
		return b
	}
	b, err := c.newSemaasBundler(namespace, kind)
	if err != nil {
		log.Printf("error creating bundler: %s", err)
		return nil
	}
	c.semaasBundlers[key] = b
	return b
}

func omegaBundler(sc *client.Client) (*bundler.Bundler, error) {
	oc, err := omega.New(sc)
	if err != nil {
		return nil, fmt.Errorf("error creating omega client: %s", err)
	}
	return bundler.NewBundler(&omega.LogEntry{}, func(entries interface{}) {
		bulk := entries.([]*omega.LogEntry)
		for {
			if err := oc.Load(context.Background(), bulk); err != nil {
				switch err := err.(type) {
				case *omega.LoadError:
					log.Printf("error loading entries: %s", err.Err.Error())
					for _, item := range err.InvalidEntities {
						if item.Position < len(bulk) { // just to check
							log.Printf("item %+v has error: %v", bulk[item.Position], item.Error)
						}
					}
					return
				case *omega.Err:
					log.Printf("error loading entries: %s", err)
					return
				default:
					log.Printf("error loading entries to omega: %s. Waiting to re-upload.", err)
					time.Sleep(5 * time.Second)
				}
			}
			return
		}
	}), nil
}

func rhoBundler(sc *client.Client) (*bundler.Bundler, error) {
	rc, err := rho.New(sc)
	if err != nil {
		return nil, fmt.Errorf("error creating rho client: %s", err)
	}
	return bundler.NewBundler(&rho.Span{}, func(entries interface{}) {
		bulk := entries.([]*rho.Span)
		for {
			if err := rc.Create(context.Background(), bulk); err != nil {
				switch err := err.(type) {
				case *omega.LoadError:
					log.Printf("error loading entries: %s", err.Err.Error())
					for _, item := range err.InvalidEntities {
						if item.Position < len(bulk) { // just to check
							log.Printf("item %v has error: %v", bulk[item.Position], item.Error)
						}
					}
					return
				case *omega.Err:
					log.Printf("error loading entries: %s", err)
					return
				default:
					log.Printf("error loading entries to omega: %s. Waiting to re-upload.", err)
					time.Sleep(5 * time.Second)
				}
			}
			return
		}
	}), nil
}

func (c *semmasClient) newSemaasBundler(namespace, kind string) (*bundler.Bundler, error) {
	var b *bundler.Bundler
	switch kind {
	case kindLogs:
		opts := omega.EnvOptions()
		opts = append(opts,
			client.WithClientCert(c.cert),
			client.WithNamespace(namespace),
			client.WithSkipVerify(),
		)
		if c.omegaURL != "" {
			opts = append(opts, client.WithURL(c.omegaURL))
		}
		sc, err := client.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("error creating semaas client: %s", err)
		}
		b, err = omegaBundler(sc)
		if err != nil {
			return nil, err
		}
	case kindSpans:
		opts := rho.EnvOptions()
		opts = append(opts,
			client.WithClientCert(c.cert),
			client.WithNamespace(namespace),
			client.WithSkipVerify(),
		)
		if c.rhoURL != "" {
			opts = append(opts, client.WithURL(c.rhoURL))
		}
		sc, err := client.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("error creating semaas client: %s", err)
		}
		b, err = rhoBundler(sc)
		if err != nil {
			return nil, err
		}
	}

	b.DelayThreshold = DefaultDelayThreshold
	b.BundleCountThreshold = DefaultEntryCountThreshold
	b.BundleByteThreshold = DefaultEntryByteThreshold
	b.BufferedByteLimit = DefaultBufferedByteLimit
	return b, nil
}

func newMuClient(muURL string, cert tls.Certificate, timeout time.Duration) (*mu.Client, error) {
	opts := mu.EnvOptions()
	opts = append(opts,
		client.WithClientCert(cert),
		client.WithNamespace("EMPTY"),
		client.WithSkipVerify(),
		client.WithTimeout(timeout),
	)
	if muURL != "" {
		opts = append(opts, client.WithURL(muURL))
	}
	sc, err := client.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating semaas client: %s", err)
	}
	mc, err := mu.New(sc)
	if err != nil {
		return nil, fmt.Errorf("error creating rho client: %s", err)
	}
	return mc, nil
}

func (c *semmasClient) processMetricV1(event *beat.Event, namespace, fallbackMrID, kind string) error {
	metric := mu.Metrics{Timestamp: event.Timestamp}

	if values, err := event.GetValue("semaas.metric.values"); err == nil {
		if valuesM, ok := values.(map[string]mu.MetricValue); ok {
			metric.Values = valuesM
		} else {
			return nil
		}
	}

	if properties, err := event.GetValue("semaas.properties"); err == nil {
		var ok bool
		metric.Properties, ok = properties.(map[string]interface{})
		if !ok {
			props := properties.(common.MapStr)
			metric.Properties = props
		}
	}
	if metric.Properties == nil {
		metric.Properties = make(map[string]interface{})
	}
	if kind == "metricv1" {
		for _, field := range c.additionalPropertyFields {
			if properties, err := event.GetValue(field); err == nil {
				for k, v := range properties.(common.MapStr) {
					if vc, ok := v.(string); ok {
						metric.Properties[k] = vc
					}
				}
			}
		}
	}

	var metricSetID string
	if msID, err := event.GetValue("semaas.metric.metricSetId"); err == nil {
		metricSetID = msID.(string)
	}

	if metricSetID != "" {
		if err := c.muClient.AddMeasurements(metricSetID, []mu.Metrics{metric}, mu.WithNamespace(namespace)); err != nil {
			c.log.Errorf("Error sending metric: %v\n", err)
		}
	}
	return nil
}
