package deepseek

import (
	"context"
	"net/http"
	"time"

	"ds2api/internal/auth"
	"ds2api/internal/config"
	trans "ds2api/internal/deepseek/transport"
	"ds2api/internal/devcapture"
	"ds2api/internal/util"
)

// intFrom is a package-internal alias for the shared util version.
var intFrom = util.IntFrom

type Client struct {
	Store      *config.Store
	Auth       *auth.Resolver
	capture    *devcapture.Store
	regular    trans.Doer
	stream     trans.Doer
	fallback   *http.Client
	fallbackS  *http.Client
	powSolver  *PowSolver
	maxRetries int
}

type accountClients struct {
	regular  trans.Doer
	fallback *http.Client
}

type proxyDoer struct {
	client *http.Client
}

func (p *proxyDoer) Do(req *http.Request) (*http.Response, error) {
	return p.client.Do(req)
}

func NewClient(store *config.Store, resolver *auth.Resolver) *Client {
	return &Client{
		Store:      store,
		Auth:       resolver,
		capture:    devcapture.Global(),
		regular:    trans.New(60 * time.Second),
		stream:     trans.New(0),
		fallback:   &http.Client{Timeout: 60 * time.Second},
		fallbackS:  &http.Client{Timeout: 0},
		powSolver:  NewPowSolver(config.WASMPath()),
		maxRetries: 3,
	}
}

func (c *Client) PreloadPow(ctx context.Context) error {
	return nil
}

func (c *Client) requestClientsForAccount(proxyURL string) accountClients {
	if proxyURL == "" {
		return accountClients{regular: c.regular, fallback: c.fallback}
	}
	proxyTransport := trans.NewWithProxy(60*time.Second, proxyURL)
	proxyClient := &http.Client{Timeout: 60 * time.Second, Transport: proxyTransport}
	return accountClients{regular: &proxyDoer{client: proxyClient}, fallback: proxyClient}
}
