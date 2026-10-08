package handlers

import (
	"context"
	"net/http"
	"time"
)

type GetHealthHandler struct{ *Handler }

func NewGetHealthHandler(h *Handler) *GetHealthHandler { return &GetHealthHandler{h} }

func (h *GetHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.Ctl.Ping(ctx); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}
