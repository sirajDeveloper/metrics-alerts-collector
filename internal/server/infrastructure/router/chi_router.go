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
}

func NewChiRouter(metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter) *ChiRouter {
	r := chi.NewRouter()

	r.Use(chiMidware.Recoverer)
	r.Use(customMidWare.LoggingMiddleware)

	handler := httpHandler.NewMetricsHandler(metricUpdater, metricGetter)

	r.Get("/", handler.GetAllMetrics)
	r.Post("/update", handler.UpdateMetric)
	r.Get("/value", handler.GetMetricValue)

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricURLParam)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handler.GetMetricValueURLParam)
	})

	return &ChiRouter{
		router:         r,
		metricsHandler: handler,
	}
}

func (cr *ChiRouter) Handler() http.Handler {
	return cr.router
}
