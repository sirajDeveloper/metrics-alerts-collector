package router

import (
	"crypto/rsa"
	"net/http"

	customMidWare "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http/middleware"

	"github.com/go-chi/chi/v5"
	chiMidware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	httpHandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type ChiRouter struct {
	router         chi.Router
	metricsHandler *httpHandler.MetricsHandler
	healthHandler  *httpHandler.HealthHandler
	secretKey      string
	trustedSubnet  *string
}

func NewChiRouter(
	metricUpdater usecase.MetricUpdater,
	metricGetter usecase.MetricGetter,
	healthChecker usecase.HealthChecker,
	secretKey string,
	auditPublisher event.AuditEventPublisher,
	privateKey *rsa.PrivateKey,
	trustedSubnet *string) *ChiRouter {
	r := chi.NewRouter()

	r.Use(chiMidware.Recoverer)
	if trustedSubnet != nil && *trustedSubnet != "" {
		r.Use(customMidWare.TrustedSubnetMiddleware(*trustedSubnet))
	}
	r.Use(customMidWare.LoggingMiddleware)
	r.Use(customMidWare.GzipMiddleware)
	if privateKey != nil {
		r.Use(customMidWare.RequestDecrypt(privateKey))
	}
	if secretKey != "" {
		r.Use(customMidWare.RequestSignatureCheck(secretKey))
		r.Use(customMidWare.ResponseSignatureAdd(secretKey))
	}

	handler := httpHandler.NewMetricsHandler(metricUpdater, metricGetter)
	healthHandler := httpHandler.NewHealthHandler(healthChecker)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

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
		trustedSubnet:  trustedSubnet,
	}
}

func (cr *ChiRouter) Handler() http.Handler {
	return cr.router
}
