package router

import (
	"net/http"

	customMidWare "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http/middleware"

	"github.com/go-chi/chi/v5"
	chiMidware "github.com/go-chi/chi/v5/middleware"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	httpHandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type ChiRouter struct {
	router         chi.Router
	metricsHandler *httpHandler.MetricsHandler
	healthHandler  *httpHandler.HealthHandler
	secretKey      string
}

func NewChiRouter(metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter, healthChecker usecase.HealthChecker, secretKey string, auditPublisher event.AuditEventPublisher) *ChiRouter {
	r := chi.NewRouter()

	r.Use(chiMidware.Recoverer)
	r.Use(customMidWare.LoggingMiddleware)
	r.Use(customMidWare.GzipMiddleware)
	if secretKey != "" {
		r.Use(customMidWare.RequestSignatureCheck(secretKey))
		r.Use(customMidWare.ResponseSignatureAdd(secretKey))
	}

	handler := httpHandler.NewMetricsHandler(metricUpdater, metricGetter, auditPublisher)
	healthHandler := httpHandler.NewHealthHandler(healthChecker)

	r.Get("/", handler.GetAllMetrics)
	r.Get("/ping", healthHandler.Ping)

	r.Route("/update", func(r chi.Router) {
		r.Post("/", handler.UpdateMetric)
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricURLParam)
	})

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", handler.UpdateMetrics)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", handler.GetMetricValue)
		r.Get("/{type}/{name}", handler.GetMetricValueURLParam)
	})

	return &ChiRouter{
		router:         r,
		metricsHandler: handler,
		healthHandler:  healthHandler,
	}
}

func (cr *ChiRouter) Handler() http.Handler {
	return cr.router
}
