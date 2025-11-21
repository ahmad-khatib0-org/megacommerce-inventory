// Package server binds everything required together for this service,
// E,g grpc, init metrics, oauth server, listen to errors, init clients....
package server

import (
	"context"
	"sync"

	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/common"
	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/controller"
	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/store"
	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/store/dbstore"
	intModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/logger"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	commonClient   *common.CommonClient
	configMux      sync.RWMutex
	configFn       func() *com.Config
	config         *com.Config
	errors         chan *models.InternalError
	tracerProvider *sdktrace.TracerProvider
	metrics        *grpcprom.ServerMetrics
	log            *logger.Logger
	dbConn         *pgxpool.Pool
	dbStore        store.InventoryDBStore
}

type ServerArgs struct {
	Log *logger.Logger
	Cfg *intModels.Config
}

// TODO: cleanup on bootstrap errors
func RunServer(s *ServerArgs) error {
	com, err := common.NewCommonClient(&common.CommonArgs{Config: s.Cfg, Log: s.Log})
	srv := &Server{
		commonClient: com,
		errors:       make(chan *models.InternalError, 1),
		log:          s.Log,
	}
	if err != nil {
		srv.errors <- err
	}

	srv.initSharedConfig()
	srv.initTrans()
	srv.initDB()
	srv.dbStore = dbstore.NewInventoryStore(srv.dbConn)

	_, err = controller.NewController(&controller.ControllerArgs{
		Config:         srv.configFn,
		TracerProvider: srv.tracerProvider,
		Metrics:        srv.metrics,
		Log:            srv.log,
	})
	if err != nil {
		srv.errors <- err
	}

	err = <-srv.errors
	if err != nil {
		s.Log.Infof("an error occurred %v ", err)
	}
	srv.shutdown()
	return err
}

func (s *Server) shutdown() {
	ctx := context.Background()
	if s.dbConn != nil {
		s.dbConn.Close()
	}

	if s.tracerProvider != nil {
		if err := s.tracerProvider.Shutdown(ctx); err != nil {
			s.log.Errorf("failed to shutdown tracer provider %v", err)
		}
	}
}
