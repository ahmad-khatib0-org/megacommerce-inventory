// Package common connect to the common service for trans, config ...
package common

import (
	"context"
	"fmt"
	"net"
	"time"

	internalModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	shared "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/shared/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/logger"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CommonClient struct {
	com.UnimplementedCommonServiceServer
	cfg    *internalModels.Config
	conn   *grpc.ClientConn
	client com.CommonServiceClient
	log    *logger.Logger
}

type CommonArgs struct {
	Config *internalModels.Config
	Log    *logger.Logger
}

// NewCommonClient runs the CommonService client
func NewCommonClient(ca *CommonArgs) (*CommonClient, *models.InternalError) {
	c := &CommonClient{cfg: ca.Config, log: ca.Log}
	if err := c.initCommonClient(); err != nil {
		return nil, err
	}

	return c, nil
}

func (cc *CommonClient) initCommonClient() *models.InternalError {
	target := cc.cfg.Service.CommonServiceGrpcURL

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "user.common.initCommonClient", Err: err, Msg: msg}
	}

	if _, err := net.ResolveTCPAddr("tcp", target); err != nil {
		return ie(err, "failed to init common service client, invalid grpc url")
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return ie(err, "failed to connect to the shared common service")
	}

	fmt.Println("user service connected to common service at: " + target)
	cc.client = com.NewCommonServiceClient(conn)
	cc.conn = conn

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cc.client.Ping(ctx, &shared.PingRequest{})
	if err != nil {
		return ie(err, "failed to ping the common service")
	}

	return nil
}

func (cc *CommonClient) Close() error {
	return cc.conn.Close()
}

func (cc *CommonClient) Conn() *grpc.ClientConn {
	return cc.conn
}
