package controller

import (
	"context"
	"time"

	intModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"google.golang.org/grpc/codes"
)

// InventoryGet returns a single inventory item by ID
func (c *Controller) InventoryGet(ctx context.Context, req *pb.InventoryGetRequest) (*pb.InventoryGetResponse, error) {
	startTime := time.Now()
	path := "inventory.controller.InventoryGet"
	modelsCtx, ctxErr := models.ContextGet(ctx)

	errBuilder := func(e *models.AppError) (*pb.InventoryGetResponse, error) {
		duration := time.Since(startTime).Seconds()
		c.metricsCollector.RecordInventoryGetRequest(false, duration)
		return &pb.InventoryGetResponse{Response: &pb.InventoryGetResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}
	if ctxErr != nil {
		return errBuilder(ctxErr)
	}
	internalErr := func(err error, details string) (*pb.InventoryGetResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}))
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	ar := models.AuditRecordNew(modelsCtx, intModels.EventNameInventoryGet, models.EventStatusFail)
	defer c.ProcessAudit(ar)

	userID := modelsCtx.Session.UserID
	if userID == "" {
		return errBuilder(models.NewAppError(modelsCtx, path, "error.unauthenticated", nil, "user not authenticated", int(codes.Unauthenticated), nil))
	}

	item, err := c.store.InventoryItemGetByID(modelsCtx, userID, req.GetId())
	if err != nil {
		if err.ErrType == models.DBErrorTypeNoRows {
			return errBuilder(models.NewAppError(modelsCtx, path, "error.not_found", nil, "", int(codes.NotFound), nil))
		}
		return internalErr(err, "failed to get inventory item")
	}
	ar.Success()

	duration := time.Since(startTime).Seconds()
	c.metricsCollector.RecordInventoryGetRequest(true, duration)

	return &pb.InventoryGetResponse{
		Response: &pb.InventoryGetResponse_Data{Data: &pb.InventoryGetResponseData{Item: item}},
	}, nil
}
