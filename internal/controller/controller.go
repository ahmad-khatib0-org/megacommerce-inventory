// Package controller contains the grpc handlers for this service
package controller

import (
	"net"
	"net/http"

	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/otel"
	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/store"
	common "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/logger"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Controller struct {
	pb.UnimplementedInventoryServiceServer
	config           func() *common.Config
	tracerProvider   *sdktrace.TracerProvider
	metrics          *grpcprom.ServerMetrics
	metricsCollector *MetricsCollector
	log              *logger.Logger
	http             *http.Client
	store            store.InventoryDBStore
}

type ControllerArgs struct {
	Config         func() *common.Config
	TracerProvider *sdktrace.TracerProvider
	Metrics        *grpcprom.ServerMetrics
	Log            *logger.Logger
	DBStore        store.InventoryDBStore
}

func NewController(ca *ControllerArgs) (*Controller, *models.InternalError) {
	// Initialize OpenTelemetry
	if _, _, err := otel.InitOTEL("megacommerce-inventory"); err != nil {
		return nil, &models.InternalError{Path: "inventory.controller.NewController", Err: err, Msg: "failed to initialize OTEL"}
	}
	// Setup Prometheus metrics endpoint on port 8065
	otel.SetupPrometheusMetrics("8065")

	c := &Controller{
		config:           ca.Config,
		tracerProvider:   ca.TracerProvider,
		metrics:          ca.Metrics,
		metricsCollector: NewMetricsCollector(),
		log:              ca.Log,
		store:            ca.DBStore,
	}

	c.http = utils.GetHTTPClient()

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "inventory.controller.NewController", Err: err, Msg: msg}
	}

	defaultLang := c.config().Localization.GetDefaultClientLocale()
	availableLangs := c.config().GetLocalization().GetAvailableLocales()
	msgSize := c.config().Services.GetInventoryServiceMaxReceiveMessageSizeBytes()

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(int(msgSize)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			models.ResponseInterceptor(defaultLang, availableLangs),
			models.UnaryMetadataInterceptor(defaultLang, availableLangs),
		),
		grpc.ChainStreamInterceptor(
			models.StreamMetadataInterceptor(defaultLang, availableLangs),
		),
	)

	addr := c.config().GetServices().GetInventoryServiceGrpcUrl()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, ie(err, "failed to initiate an http listener")
	}

	reflection.Register(s)
	pb.RegisterInventoryServiceServer(s, c)
	// c.metrics.InitializeMetrics(s)

	go func() {
		c.log.Infof("grpc inventory service is running on %s", addr)
		if err := s.Serve(listener); err != nil {
			s.GracefulStop()
			s.Stop()
		}
	}()

	return c, nil
}
