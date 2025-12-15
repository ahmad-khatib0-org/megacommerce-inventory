package controller

import (
	"context"
	"time"

	intModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"google.golang.org/grpc/codes"
)

// InventoryList returns a paginated list of inventory items for the authenticated supplier
func (c *Controller) InventoryList(ctx context.Context, req *pb.InventoryListRequest) (*pb.InventoryListResponse, error) {
	path := "inventory.controller.InventoryList"
	modelsCtx, ctxErr := models.ContextGet(ctx)

	errBuilder := func(e *models.AppError) (*pb.InventoryListResponse, error) {
		return &pb.InventoryListResponse{Response: &pb.InventoryListResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}

	if ctxErr != nil {
		return errBuilder(ctxErr)
	}

	internalErr := func(err error, details string) (*pb.InventoryListResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}))
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	ar := models.AuditRecordNew(modelsCtx, intModels.EventNameInventoryList, models.EventStatusFail)
	defer c.ProcessAudit(ar)

	userID := modelsCtx.Session.UserID
	if userID == "" {
		return errBuilder(models.NewAppError(modelsCtx, path, "error.unauthenticated", nil, "user not authenticated", int(codes.Unauthenticated), nil))
	}

	pagination := req.GetPagination()

	if err := models.CheckLastID(modelsCtx, "inventory.controller.InventoryList", pagination); err != nil {
		return errBuilder(err)
	}

	pageSize := int32(20) // Default page size
	// if pagination.PageSize != nil && *pagination.PageSize > 0 {
	// 	pageSize = int32(*pagination.PageSize)
	// }

	lastID := ""
	if pagination.LastId != nil {
		lastID = *pagination.LastId
	}

	// Query inventory items for this supplier using store layer
	items, err := c.store.InventoryList(modelsCtx, userID, pageSize+1, lastID)
	if err != nil {
		return internalErr(err, "failed to query inventory items")
	}

	// Check if we have more items than pageSize to determine hasNext
	if int32(len(items)) > pageSize {
		items = items[:pageSize]
	}

	ar.Success()

	return &pb.InventoryListResponse{
		Response: &pb.InventoryListResponse_Data{
			Data: &pb.InventoryListResponseData{
				Items:      items,
				Pagination: models.BuildPaginationResponse(pagination, len(items)),
			},
		},
	}, nil
}
