package traefik_svix_plugin

import (
	"context"
	svix "github.com/svix/svix-webhooks/go"
	"io"
	"net/http"
)

type Config struct {
	SvixSigningSecret string
}

func CreateConfig() *Config {
	return &Config{
		SvixSigningSecret: "",
	}
}

type SvixPlugin struct {
	next http.Handler
	name string
	wh   *svix.Webhook
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	wh, err := svix.NewWebhook(config.SvixSigningSecret)

	if err != nil {
		return nil, err
	}

	return &SvixPlugin{
		next: next,
		name: name,
		wh:   wh,
	}, nil
}

func (s *SvixPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	payload, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	err = s.wh.Verify(payload, req.Header)

	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	s.next.ServeHTTP(rw, req)
}
