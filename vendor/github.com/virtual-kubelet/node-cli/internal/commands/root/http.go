// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package root

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/virtual-kubelet/node-cli/opts"
	"github.com/virtual-kubelet/node-cli/provider"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
)

// AcceptedCiphers is the list of accepted TLS ciphers, with known weak ciphers elided
// Note this list should be a moving target.
var AcceptedCiphers = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,

	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

func loadTLSConfig(ctx context.Context, certPath, keyPath, caPath string, allowUnauthenticatedClients, authWebhookEnabled bool) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading tls certs")
	}

	var (
		caPool     *x509.CertPool
		clientAuth = tls.RequireAndVerifyClientCert
	)

	if allowUnauthenticatedClients {
		clientAuth = tls.NoClientCert
	}
	if authWebhookEnabled {
		clientAuth = tls.RequestClientCert
	}

	if caPath != "" {
		caPool = x509.NewCertPool()
		pem, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		if !caPool.AppendCertsFromPEM(pem) {
			return nil, errors.New("error appending ca cert to certificate pool")
		}
	}

	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             AcceptedCiphers,
		ClientCAs:                caPool,
		ClientAuth:               clientAuth,
	}, nil
}

func setupHTTPServer(ctx context.Context, p provider.Provider, cfg *apiServerConfig) (_ func(), retErr error) {
	var closers []io.Closer
	cancel := func() {
		for _, c := range closers {
			c.Close()
		}
	}
	defer func() {
		if retErr != nil {
			cancel()
		}
	}()

	if cfg.CertPath == "" || cfg.KeyPath == "" || (cfg.CACertPath == "" && !cfg.AllowUnauthenticatedClients) {
		log.G(ctx).
			WithField("certPath", cfg.CertPath).
			WithField("keyPath", cfg.KeyPath).
			WithField("caPath", cfg.CACertPath).
			Error("TLS certificates not provided, not setting up pod http server")
	} else {
		tlsCfg, err := loadTLSConfig(ctx, cfg.CertPath, cfg.KeyPath, cfg.CACertPath, cfg.AllowUnauthenticatedClients, cfg.AuthWebhookEnabled)
		if err != nil {
			return nil, err
		}
		l, err := tls.Listen("tcp", cfg.Addr, tlsCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "error setting up listener for pod http server: tlsconfig: \n%+v", tlsCfg)
		}

		mux := NewServeMuxWithAuth(ctx, cfg.Auth)

		podRoutes := api.PodHandlerConfig{
			RunInContainer:        p.RunInContainer,
			GetContainerLogs:      p.GetContainerLogs,
			GetPods:               p.GetPods,
			StreamIdleTimeout:     cfg.StreamIdleTimeout,
			StreamCreationTimeout: cfg.StreamCreationTimeout,
		}

		if mp, ok := p.(provider.PodMetricsProvider); ok {
			podRoutes.GetStatsSummary = mp.GetStatsSummary
		}

		api.AttachPodRoutes(podRoutes, mux, true)

		s := &http.Server{
			Handler:   mux,
			TLSConfig: tlsCfg,
		}
		go serveHTTP(ctx, s, l, "pods")
		closers = append(closers, s)
	}

	if cfg.MetricsAddr != "" {
		l, err := net.Listen("tcp", cfg.MetricsAddr)
		if err != nil {
			return nil, errors.Wrap(err, "could not setup listener for pod metrics http server")
		}
		var summaryHandlerFunc api.PodStatsSummaryHandlerFunc
		if mp, ok := p.(provider.PodMetricsProvider); ok {
			summaryHandlerFunc = mp.GetStatsSummary
		}
		podMetricsRoutes := api.PodMetricsConfig{
			GetStatsSummary: summaryHandlerFunc,
		}

		mux := http.NewServeMux()
		api.AttachPodMetricsRoutes(podMetricsRoutes, mux)
		s := &http.Server{
			Handler: mux,
		}
		go serveHTTP(ctx, s, l, "pod metrics")
		closers = append(closers, s)
	}

	return cancel, nil
}

func serveHTTP(ctx context.Context, s *http.Server, l net.Listener, name string) {
	if err := s.Serve(l); err != nil {
		select {
		case <-ctx.Done():
		default:
			log.G(ctx).WithError(err).Errorf("Error setting up %s http server", name)
		}
	}
	l.Close()
}

type apiServerConfig struct {
	CACertPath                  string
	CertPath                    string
	KeyPath                     string
	Addr                        string
	MetricsAddr                 string
	StreamIdleTimeout           time.Duration
	StreamCreationTimeout       time.Duration
	AllowUnauthenticatedClients bool

	Auth               AuthInterface
	AuthWebhookEnabled bool
}

func getAPIConfig(c *opts.Opts) (*apiServerConfig, error) {
	config := apiServerConfig{
		CertPath: os.Getenv("APISERVER_CERT_LOCATION"),
		KeyPath:  os.Getenv("APISERVER_KEY_LOCATION"),
	}

	config.AuthWebhookEnabled = c.Authentication.Webhook.Enabled
	config.Addr = fmt.Sprintf(":%d", c.ListenPort)
	config.MetricsAddr = c.MetricsAddr
	config.StreamIdleTimeout = c.StreamIdleTimeout
	config.StreamCreationTimeout = c.StreamCreationTimeout
	config.AllowUnauthenticatedClients = c.AllowUnauthenticatedClients

	config.CACertPath = c.ClientCACert
	if c.ClientCACert == "" {
		config.CACertPath = os.Getenv("APISERVER_CA_CERT_LOCATION")
	}

	return &config, nil
}
