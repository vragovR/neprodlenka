package bot

import (
	"crypto/subtle"
	"net/http"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"

	"neprodlenka/internal/config"
	"neprodlenka/internal/maxapi"
)

type Webhook struct {
	secret  string
	handler http.Handler
}

func NewWebhook(api *maxapi.Client, b *Bot, cfg config.Config) *Webhook {
	return &Webhook{
		secret:  cfg.WebhookSecret,
		handler: api.WebhookHandler(cfg.WebhookSecret, b.HandleUpdate),
	}
}

func (wh *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if wh.secret == "" {
		http.NotFound(w, r)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get(maxbot.SecretHeader)), []byte(wh.secret)) != 1 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	wh.handler.ServeHTTP(w, r)
}
