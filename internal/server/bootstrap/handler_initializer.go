package bootstrap

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/pkg/crypto"
)

type HandlerInitializer struct {
	config         Config
	metricUpdater  usecase.MetricUpdater
	metricGetter   usecase.MetricGetter
	healthService  usecase.HealthChecker
	emitter        *usecase.MetricsEmitterService
	auditPublisher event.AuditEventPublisher
}

func NewHandlerInitializer(cfg Config, metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter, healthService usecase.HealthChecker, emitter *usecase.MetricsEmitterService, auditPublisher event.AuditEventPublisher) *HandlerInitializer {
	return &HandlerInitializer{
		config:         cfg,
		metricUpdater:  metricUpdater,
		metricGetter:   metricGetter,
		healthService:  healthService,
		emitter:        emitter,
		auditPublisher: auditPublisher,
	}
}

type HandlerResult struct {
	Server      *http.Server
	Scheduler   *scheduler.MetricEmitterScheduler
	SchedCtx    context.Context
	SchedCancel context.CancelFunc
	EnableHTTPS bool
	TLSCertFile string
	TLSKeyFile  string
}

func (h *HandlerInitializer) Initialize() *HandlerResult {
	schedulerInstance := scheduler.NewMetricEmitterScheduler(h.emitter, *h.config.GetStoreInterval(), *h.config.GetRestore())
	schedCtx, schedCancel := context.WithCancel(context.Background())
	schedulerInstance.Start(schedCtx)

	var privateKey *rsa.PrivateKey
	if h.config.GetCryptoKey() != nil && *h.config.GetCryptoKey() != "" {
		key, err := crypto.LoadPrivateKey(*h.config.GetCryptoKey())
		if err != nil {
			log.Fatalf("Failed to load private key: %v", err)
		}
		privateKey = key
		log.Printf("Private key loaded from: %s", *h.config.GetCryptoKey())
	}

	secretKey := ""
	if h.config.GetSecretKey() != nil {
		secretKey = *h.config.GetSecretKey()
	}

	chiRouter := router.NewChiRouter(h.metricUpdater, h.metricGetter, h.healthService, secretKey, h.auditPublisher, privateKey)
	server := &http.Server{
		Addr:    *h.config.GetAddress(),
		Handler: chiRouter.Handler(),
	}

	enableHTTPS := false
	tlsCertFile := ""
	tlsKeyFile := ""

	if h.config.GetEnableHTTPS() != nil && *h.config.GetEnableHTTPS() {
		enableHTTPS = true
		if h.config.GetTLSCertFile() != nil {
			tlsCertFile = *h.config.GetTLSCertFile()
		}
		if h.config.GetTLSKeyFile() != nil {
			tlsKeyFile = *h.config.GetTLSKeyFile()
		}
	}

	return &HandlerResult{
		Server:      server,
		Scheduler:   schedulerInstance,
		SchedCtx:    schedCtx,
		SchedCancel: schedCancel,
		EnableHTTPS: enableHTTPS,
		TLSCertFile: tlsCertFile,
		TLSKeyFile:  tlsKeyFile,
	}
}
