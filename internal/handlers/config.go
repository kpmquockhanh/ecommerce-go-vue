package handlers

import (
	"net/http"
)

type ConfigHandler struct {
	stripePublishableKey string
}

func NewConfigHandler(stripePublishableKey string) *ConfigHandler {
	return &ConfigHandler{stripePublishableKey: stripePublishableKey}
}

func (h *ConfigHandler) GetPublicConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondWithJSON(w, map[string]string{
		"stripe_publishable_key": h.stripePublishableKey,
	}, http.StatusOK)
}
