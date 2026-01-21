package http

import (
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type HealthHandler struct {
	healthChecker usecase.HealthChecker
}

func NewHealthHandler(healthChecker usecase.HealthChecker) *HealthHandler {
	return &HealthHandler{healthChecker: healthChecker}
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.healthChecker == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := h.healthChecker.Ping(r.Context()); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
