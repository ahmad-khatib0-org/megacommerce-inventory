package common

import (
	"context"
	"fmt"
	"time"

	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
)

func (cc *CommonClient) ConfigGet() (*com.Config, *models.InternalError) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "user.common.ConfigGet", Err: err, Msg: msg}
	}

	res, err := cc.client.ConfigGet(ctx, &com.ConfigGetRequest{})
	if err != nil {
		return nil, ie(err, "failed to get configurations from common service")
	}

	switch res := res.Response.(type) {
	case *com.ConfigGetResponse_Data:
		return res.Data, nil
	case *com.ConfigGetResponse_Error:
		err := models.AppErrorFromProto(nil, res.Error) // no need for ctx here
		return nil, ie(err, "failed to get configurations from common service")
	}

	return nil, nil
}

// TODO: complete this listener
func (cc *CommonClient) ConfigListener(clientID string) *models.InternalError {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "user.common.ConfigListener", Err: err, Msg: msg}
	}

	stream, err := cc.client.ConfigListener(ctx, &com.ConfigListenerRequest{ClientId: clientID})
	if err != nil {
		return ie(err, "failed to register a listener call")
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			fmt.Println("config listener stream closed")
			break
		}

		switch x := res.Response.(type) {
		case *com.ConfigListenerResponse_Data:
			fmt.Println("Config changed: ", x.Data)
		case *com.ConfigListenerResponse_Error:
			fmt.Println("Error received: ", x.Error.Message)
		}
	}

	return nil
}
