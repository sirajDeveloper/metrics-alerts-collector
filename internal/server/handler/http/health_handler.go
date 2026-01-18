package http

import (
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

// HealthHandler обрабатывает запросы для проверки здоровья сервиса.
// Эндпоинт: GET /ping
type HealthHandler struct {
	healthChecker usecase.HealthChecker
}

// NewHealthHandler создает новый экземпляр HealthHandler.
//
// Параметры:
//   - healthChecker: интерфейс для проверки здоровья (может быть nil)
//
// Возвращает новый экземпляр HealthHandler.
func NewHealthHandler(healthChecker usecase.HealthChecker) *HealthHandler {
	return &HealthHandler{healthChecker: healthChecker}
}

// Ping обрабатывает GET /ping.
// Проверяет доступность базы данных и возвращает "OK" при успехе.
//
// Ответы:
//   - 200 OK: база данных доступна, возвращает "OK"
//   - 500 Internal Server Error: база данных недоступна или не настроена
//
// @Summary Проверка здоровья
// @Description Проверяет доступность базы данных
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "OK"
// @Failure 500 "База данных недоступна"
// @Router /ping [get]
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
