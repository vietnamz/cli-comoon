package client

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
)

// Opt is a configuration option to initialize a client
type Opt func(*Client) error

// FromEnv configures the client with values from environment variables.
//
// Supported environment variables:
// HOST to set the url to the docker server.
// API_VERSION to set the version of the API to reach, leave empty for latest.
// CERT_PATH to load the TLS certificates from.
// TLS_VERIFY to enable or disable TLS verification, off by default.
func FromEnv(c *Client) error {
	if dockerCertPath := os.Getenv("CERT_PATH"); dockerCertPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
			CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
			KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
			InsecureSkipVerify: os.Getenv("TLS_VERIFY") == "",
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return err
		}

		c.client = &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: CheckRedirect,
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		if err := WithHost(host)(c); err != nil {
			return err
		}
	}

	if version := os.Getenv("API_VERSION"); version != "" {
		if err := WithVersion(version)(c); err != nil {
			return err
		}
	}
	return nil
}

// WithDialer applies the dialer.DialContext to the client transport. This can be
// used to set the Timeout and KeepAlive settings of the client.
// Deprecated: use WithDialContext
func WithDialer(dialer *net.Dialer) Opt {
	return WithDialContext(dialer.DialContext)
}

// WithDialContext applies the dialer to the client transport. This can be
// used to set the Timeout and KeepAlive settings of the client.
func WithDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) Opt {
	return func(c *Client) error {
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.DialContext = dialContext
			return nil
		}
		return errors.Errorf("cannot apply dialer to transport: %T", c.client.Transport)
	}
}

// WithHost overrides the client host with the specified one.
func WithHost(host string) Opt {
	return func(c *Client) error {
		hostURL, err := ParseHostURL(host)
		if err != nil {
			return err
		}
		c.host = host
		c.proto = hostURL.Scheme

		/* TODO:
				please investigate why I have to change c.scheme.
				Do not understand why for now. But it works. :))
		 */
		c.scheme = hostURL.Scheme
		//--------------------------------------------------------
		c.addr = hostURL.Host
		c.basePath = hostURL.Path
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			logrus.Errorf("Configure transport proto %s, addr %s", c.proto, c.addr)
			err = sockets.ConfigureTransport(transport, c.proto, c.addr)
			if err != nil {
				logrus.Errorf("Configure transport err %s", err)
			}
			return nil
		}
		return errors.Errorf("cannot apply host to transport: %T", c.client.Transport)
	}
}

// WithHTTPClient overrides the client http client with the specified one
func WithHTTPClient(client *http.Client) Opt {
	return func(c *Client) error {
		if client != nil {
			c.client = client
		}
		return nil
	}
}

// WithTimeout configures the time limit for requests made by the HTTP client
func WithTimeout(timeout time.Duration) Opt {
	return func(c *Client) error {
		c.client.Timeout = timeout
		return nil
	}
}

// WithHTTPHeaders overrides the client default http headers
func WithHTTPHeaders(headers map[string]string) Opt {
	return func(c *Client) error {
		c.customHTTPHeaders = headers
		return nil
	}
}

// WithScheme overrides the client scheme with the specified one
func WithScheme(scheme string) Opt {
	return func(c *Client) error {
		c.scheme = scheme
		return nil
	}
}

// WithTLSClientConfig applies a tls config to the client transport.
func WithTLSClientConfig(cacertPath, certPath, keyPath string) Opt {
	return func(c *Client) error {
		opts := tlsconfig.Options{
			CAFile:             cacertPath,
			CertFile:           certPath,
			KeyFile:            keyPath,
			ExclusiveRootPools: true,
		}
		config, err := tlsconfig.Client(opts)
		if err != nil {
			return errors.Wrap(err, "failed to create tls config")
		}
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = config
			return nil
		}
		return errors.Errorf("cannot apply tls config to transport: %T", c.client.Transport)
	}
}

// WithVersion overrides the client version with the specified one. If an empty
// version is specified, the value will be ignored to allow version negotiation.
// With the external endpoint. We need to allow the version is different from ours.
// So we should allow override the version here to make sure we can call the correct
// endpoint.
func WithVersion(version string) Opt {
	return func(c *Client) error {
		c.version = version
		c.manualOverride = true
		return nil
	}
}

// WithAPIVersionNegotiation enables automatic API version negotiation for the client.
// With this option enabled, the client automatically negotiates the API version
// to use when making requests. API version negotiation is performed on the first
// request; subsequent requests will not re-negotiate.
func WithAPIVersionNegotiation() Opt {
	return func(c *Client) error {
		c.negotiateVersion = true
		return nil
	}
}
