// firehose-nozzle-v2
//
// Copyright (c) 2018-Present Pivotal Software, Inc. All Rights Reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may
// obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing permissions and
// limitations under the License.

package nozzle

import (
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

var allSelectors = []*loggregator_v2.Selector{
	{
		Message: &loggregator_v2.Selector_Log{
			Log: &loggregator_v2.LogSelector{},
		},
	},
	{
		Message: &loggregator_v2.Selector_Counter{
			Counter: &loggregator_v2.CounterSelector{},
		},
	},
	{
		Message: &loggregator_v2.Selector_Gauge{
			Gauge: &loggregator_v2.GaugeSelector{},
		},
	},
	{
		Message: &loggregator_v2.Selector_Timer{
			Timer: &loggregator_v2.TimerSelector{},
		},
	},
	{
		Message: &loggregator_v2.Selector_Event{
			Event: &loggregator_v2.EventSelector{},
		},
	},
}

//go:generate counterfeiter ./ TLSConfigProvider
type TLSConfigProvider interface {
	GetTLSConfig() (*tls.Config, error)
}

type tlsConfigProvider struct {
	config *Config
}

func (t *tlsConfigProvider) GetTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(t.config.CertPath, t.config.KeyPath)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ServerName:   t.config.CommonName,
		Certificates: []tls.Certificate{cert},
	}

	caCertBytes, err := ioutil.ReadFile(t.config.CACertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		return nil, errors.New("cannot parse ca cert")
	}

	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
}

func NewTLSConfig(c *Config) TLSConfigProvider {
	return &tlsConfigProvider{config: c}
}

func Receive(c *Config, tls TLSConfigProvider) error {
	tlsConfig, err := tls.GetTLSConfig()
	if err != nil {
		log.Fatal("Could not create TLS nozzle", err)
	}

	loggr := log.New(os.Stderr, "[", log.LstdFlags)

	streamConnector := loggregator.NewEnvelopeStreamConnector(
		c.LogsAPIAddr,
		tlsConfig,
		loggregator.WithEnvelopeStreamLogger(loggr),
	)

	rx := streamConnector.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{
		ShardId:   c.ShardID,
		Selectors: allSelectors,
	})

	for {
		batch := rx()

		for _, e := range batch {
			log.Printf("%+v\n", e)
		}
	}
}
