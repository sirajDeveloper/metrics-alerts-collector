package router

import (
	"net/http"

	customMidWare "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http/middleware"

	"github.com/go-chi/chi/v5"
	chiMidware "github.com/go-chi/chi/v5/middleware"

	httpHandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type ChiRouter struct {
	router         chi.Router
	metricsHandler *httpHandler.MetricsHandler
	healthHandler  *httpHandler.HealthHandler
}

func NewChiRouter(metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter, healthChecker usecase.HealthChecker) *ChiRouter {
	r := chi.NewRouter()

	r.Use(chiMidware.Recoverer)
	r.Use(customMidWare.LoggingMiddleware)
	r.Use(customMidWare.GzipMiddleware)

	handler := httpHandler.NewMetricsHandler(metricUpdater, metricGetter)
	healthHandler := httpHandler.NewHealthHandler(healthChecker)

	r.Get("/", handler.GetAllMetrics)
	r.Get("/ping", healthHandler.Ping)

	r.Route("/update", func(r chi.Router) {
		r.Post("/", handler.UpdateMetric)
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricURLParam)
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
