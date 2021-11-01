/*
 * Copyright (c) 2021 Huy Duc Dao
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package server provides tools to create http servers and handlers wrapping the turbo router
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/core"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

type ToHTTPError func(error) int

func DefaultToHTTPError(_ error) int {
	return http.StatusInternalServerError
}

const (
	HeaderCompleteResponseValue   = "true"
	HeaderIncompleteResponseValue = "false"
)

var (
	CompleteResponseHeaderName = "X-Sonic-Completed"
	HeadersToSend              = []string{"Content-Type"}
	UserAgentHeaderValue       = []string{core.SonicUserAgent}
	ErrInternalError           = errors.New("internal server error")
	ErrPrivateKey              = errors.New("private key not defined")
	ErrPublicKey               = errors.New("public key not defined")
)

func InitHTTPDefaultTransport(cfg config.ServiceConfig) {
	onceTransportConfig.Do(func() {
		http.DefaultTransport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:       cfg.DialerTimeout,
				KeepAlive:     cfg.DialerKeepAlive,
				FallbackDelay: cfg.DialerFallbackDelay,
			}).DialContext,
			DisableCompression:    cfg.DisableCompression,
			DisableKeepAlives:     cfg.DisableKeepAlives,
			MaxIdleConns:          cfg.MaxIdleConnections,
			MaxIdleConnsPerHost:   cfg.MaxIdleConnectionsPerHost,
			IdleConnTimeout:       cfg.IdleConnectionTimeout,
			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
			ExpectContinueTimeout: cfg.ExpectContinueTimeout,
			TLSHandshakeTimeout:   10 * time.Second,
		}
	})
}

func RunServer(ctx context.Context, cfg config.ServiceConfig, handler http.Handler) error {
	done := make(chan error)
	s := NewServer(cfg, handler)

	if s.TLSConfig == nil {
		go func() {
			done <- s.ListenAndServe()
		}()
	} else {
		if cfg.TLS.PublicKey == "" {
			return ErrPublicKey
		}
		if cfg.TLS.PrivateKey == "" {
			return ErrPrivateKey
		}
		go func() {
			done <- s.ListenAndServeTLS(cfg.TLS.PublicKey, cfg.TLS.PrivateKey)
		}()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

func NewServer(cfg config.ServiceConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		TLSConfig:         ParseTLSConfig(cfg.TLS),
	}
}

func ParseTLSConfig(cfg *config.TLS) *tls.Config {
	if cfg == nil {
		return nil
	}
	if cfg.IsDisabled {
		return nil
	}

	tlsConfig := &tls.Config{
		MinVersion:               parseTLSVersion(cfg.MinVersion),
		MaxVersion:               parseTLSVersion(cfg.MaxVersion),
		CurvePreferences:         parseCurveIDs(cfg),
		PreferServerCipherSuites: cfg.PreferServerCipherSuites,
		CipherSuites:             parseCipherSuites(cfg),
	}
	if !cfg.EnableMTLS {
		return tlsConfig
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		certPool = x509.NewCertPool()
	}

	caCert, err := ioutil.ReadFile(cfg.PublicKey)
	if err != nil {
		return tlsConfig
	}
	certPool.AppendCertsFromPEM(caCert)

	tlsConfig.ClientCAs = certPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	return tlsConfig
}

func parseTLSVersion(key string) uint16 {
	if v, ok := versions[key]; ok {
		return v
	}
	return tls.VersionTLS13
}

func parseCurveIDs(cfg *config.TLS) []tls.CurveID {
	l := len(cfg.CurvePreferences)
	if l == 0 {
		return defaultCurves
	}

	curves := make([]tls.CurveID, len(cfg.CurvePreferences))
	for i := range curves {
		curves[i] = tls.CurveID(cfg.CurvePreferences[i])
	}
	return curves
}

func parseCipherSuites(cfg *config.TLS) []uint16 {
	l := len(cfg.CipherSuites)
	if l == 0 {
		return defaultCipherSuites
	}

	cs := make([]uint16, l)
	for i := range cs {
		cs[i] = uint16(cfg.CipherSuites[i])
	}
	return cs
}

var (
	onceTransportConfig sync.Once
	defaultCurves       = []tls.CurveID{
		tls.CurveP521,
		tls.CurveP384,
		tls.CurveP256,
	}
	defaultCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
	versions = map[string]uint16{
		"TLS10": tls.VersionTLS10,
		"TLS11": tls.VersionTLS11,
		"TLS12": tls.VersionTLS12,
		"TLS13": tls.VersionTLS13,
	}
)
