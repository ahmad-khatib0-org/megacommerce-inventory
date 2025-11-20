// Package controller contains the grpc handlers for this service
package controller

import (
	"net"
	"net/http"

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
	config         func() *common.Config
	tracerProvider *sdktrace.TracerProvider
	metrics        *grpcprom.ServerMetrics
	log            *logger.Logger
	http           *http.Client
}

type ControllerArgs struct {
	Config         func() *common.Config
	TracerProvider *sdktrace.TracerProvider
	Metrics        *grpcprom.ServerMetrics
	Log            *logger.Logger
}

func NewController(ca *ControllerArgs) (*Controller, *models.InternalError) {
	c := &Controller{
		config:         ca.Config,
		tracerProvider: ca.TracerProvider,
		metrics:        ca.Metrics,
		log:            ca.Log,
	}

	c.http = utils.GetHTTPClient()

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "user.controller.NewController", Err: err, Msg: msg}
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
			// c.metrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(traceID)),
			// selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authMiddleware), selector.MatchFunc(authMatcher)),
		),
		grpc.ChainStreamInterceptor(
			models.StreamMetadataInterceptor(defaultLang, availableLangs),
			// c.metrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(traceID)),
			// selector.StreamServerInterceptor(auth.StreamServerInterceptor(authMiddleware), selector.MatchFunc(authMatcher)),
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
