package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"

	internalTls "server/internal/tls"

	oohttp "github.com/ooni/oohttp"
)

const (
	DefaultHttpTimeout         = time.Duration(30) * time.Second
	DefaultHttpKeepAlive       = time.Duration(30) * time.Second
	DefaultIdleConnTimeout     = time.Duration(90) * time.Second
	DefaultTLSHandshakeTimeout = time.Duration(10) * time.Second
)

type TransportConfig struct {
	// Hostname to send the HTTP request to.
	Host string

	// HTTP or HTTPs.
	Scheme string

	// The TLS fingerprint to use.
	Fingerprint internalTls.Fingerprint

	// Hexadecimal Client Hello to use
	HexClientHello internalTls.HexClientHello

	// The maximum amount of time a dial will wait for a connect to complete.
	// Defaults to [DefaultHttpTimeout].
	HttpTimeout int

	// Specifies the interval between keep-alive probes for an active network connection.
	// Defaults to [DefaultHttpKeepAlive].
	HttpKeepAliveInterval int

	// The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
	// Defaults to [DefaultIdleConnTimeout].
	IdleConnTimeout int

	// The maximum amount of time to wait for a TLS handshake.
	// Defaults to [DefaultTLSHandshakeTimeout].
	TLSHandshakeTimeout int

	// UseInterceptedFingerprint use intercepted fingerprint
	UseInterceptedFingerprint bool
}

func ParseTransportConfig(data string) (*TransportConfig, error) {
	config := &TransportConfig{}

	if strings.TrimSpace(data) == "" {
		return nil, errors.New("missing transport configuration")
	}

	if err := json.Unmarshal([]byte(data), config); err != nil {
		return nil, err
	}

	return config, nil
}

// NewTransport creates a new transport using the given configuration.
func NewTransport(config *TransportConfig, getInterceptedFingerprint func(sni string) string) (*oohttp.StdlibTransport, error) {
	dialer := &net.Dialer{
		Timeout:   DefaultHttpTimeout,
		KeepAlive: DefaultHttpKeepAlive,
	}

	if config.HttpTimeout != 0 {
		dialer.Timeout = time.Duration(config.HttpTimeout) * time.Second
	}
	if config.HttpKeepAliveInterval != 0 {
		dialer.KeepAlive = time.Duration(config.HttpKeepAliveInterval) * time.Second
	}

	var err error
	var spec *utls.ClientHelloSpec
	var clientHelloID *utls.ClientHelloID

	if config.HexClientHello != "" {
		spec, err = config.HexClientHello.ToClientHelloSpec()
		if err != nil {
			return nil, fmt.Errorf("create spec from client hello: %w", err)
		}
		clientHelloID = &utls.HelloCustom
	} else {
		clientHelloID = config.Fingerprint.ToClientHelloId()
	}

	getClientHello := func(sni string) (*utls.ClientHelloID, *utls.ClientHelloSpec) {
		interceptedFP := getInterceptedFingerprint(sni)

		if !config.UseInterceptedFingerprint || interceptedFP == "" {
			return clientHelloID, spec
		}

		interseptedSpec, err := internalTls.HexClientHello(interceptedFP).ToClientHelloSpec()
		if err == nil {
			return &utls.HelloCustom, interseptedSpec
		}

		return clientHelloID, spec
	}

	tlsFactory := &internalTls.FactoryWithClientHelloId{GetClientHello: getClientHello}

	transport := &oohttp.Transport{
		Proxy:                 oohttp.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientFactory:      tlsFactory.NewUTLSConn,
	}

	// add realistic initial HTTP2 SETTINGS to Chrome browser fingerprints
	if strings.HasPrefix(string(config.Fingerprint), "Chrome") {
		transport.EnableCustomInitialSettings()
		transport.HeaderTableSize = 4096 // 65536 // TODO: 4096 seems to be the max; modify oohtpp fork (see `http2/hpack` package) to support higher value
		transport.EnablePush = 0
		transport.MaxConcurrentStreams = 1000
		transport.InitialWindowSize = 6291456
		transport.MaxFrameSize = 16384
		transport.MaxHeaderListSize = 262144
	}

	if config.IdleConnTimeout != 0 {
		transport.IdleConnTimeout = time.Duration(config.IdleConnTimeout) * time.Second
	}
	if config.TLSHandshakeTimeout != 0 {
		transport.TLSHandshakeTimeout = time.Duration(config.TLSHandshakeTimeout) * time.Second
	}

	return &oohttp.StdlibTransport{
		Transport: transport,
	}, nil
}
