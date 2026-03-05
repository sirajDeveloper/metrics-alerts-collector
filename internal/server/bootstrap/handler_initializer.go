package bootstrap

import (
	"context"
	"crypto/rsa"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/proto"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	grpchandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/grpc"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/pkg/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	Server       *http.Server
	Scheduler    *scheduler.MetricEmitterScheduler
	SchedCtx     context.Context
	SchedCancel  context.CancelFunc
	EnableHTTPS  bool
	TLSCertFile  string
	TLSKeyFile   string
	GRPCServer   *grpc.Server
	GRPCListener net.Listener
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

	chiRouter := router.NewChiRouter(
		h.metricUpdater,
		h.metricGetter,
		h.healthService,
		secretKey,
		h.auditPublisher,
		privateKey,
		h.config.GetTrustedSubnet())
	server := &http.Server{
		Addr:    *h.config.GetAddress(),
		Handler: chiRouter.Handler(),
	}

	grpcAddr := ":3200"
	if h.config.GetGRPCAddress() != nil && *h.config.GetGRPCAddress() != "" {
		grpcAddr = *h.config.GetGRPCAddress()
	}

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen for gRPC server: %v", err)
	}
	trustedSubnet := ""
	if h.config.GetTrustedSubnet() != nil {
		trustedSubnet = *h.config.GetTrustedSubnet()
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(newTrustedSubnetInterceptor(trustedSubnet)))
	proto.RegisterMetricsServer(grpcServer, grpchandler.NewMetricsHandler(h.metricUpdater))

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
		Server:       server,
		Scheduler:    schedulerInstance,
		SchedCtx:     schedCtx,
		SchedCancel:  schedCancel,
		EnableHTTPS:  enableHTTPS,
		TLSCertFile:  tlsCertFile,
		TLSKeyFile:   tlsKeyFile,
		GRPCServer:   grpcServer,
		GRPCListener: grpcListener,
	}
}

func newTrustedSubnetInterceptor(subnet string) grpc.UnaryServerInterceptor {
	if subnet == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic("invalid subnet: " + err.Error())
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		ip := net.ParseIP(strings.TrimSpace(values[0]))
		if ip == nil || !ipNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		return handler(ctx, req)
	}
}
