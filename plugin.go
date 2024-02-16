package caddy_pirsch_plugin

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	pirsch "github.com/pirsch-analytics/pirsch-go-sdk/v2/pkg"
	"go.uber.org/zap"
	"net/http"
)

func init() {
	caddy.RegisterModule(PirschPlugin{})
}

type PirschPlugin struct {
	ClientId     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	BaseURL      string `json:"base_url,omitempty"`

	logger *zap.Logger
	client *pirsch.Client
}

func (m PirschPlugin) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.pirsch",
		New: func() caddy.Module { return new(PirschPlugin) },
	}
}

func (m *PirschPlugin) Provision(ctx caddy.Context) (err error) {
	var clientConfig *pirsch.ClientConfig
	if m.BaseURL != "" {
		clientConfig = &pirsch.ClientConfig{BaseURL: m.BaseURL}
	}

	m.client = pirsch.NewClient(m.ClientId, m.ClientSecret, clientConfig)
	m.logger = ctx.Logger(m)

	return err
}

func (m *PirschPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	r2 := r.Clone(context.TODO())
	go func(r *http.Request) {
		if err := m.client.PageView(r, nil); err != nil {
			m.logger.Error("failed to send hit to pirsch: %v", zap.Error(err))
		}
	}(r2)
	return next.ServeHTTP(w, r)
}

var _ caddyhttp.MiddlewareHandler = (*PirschPlugin)(nil)
