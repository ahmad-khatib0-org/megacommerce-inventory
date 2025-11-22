package controller

import (
	"context"
	"time"

	intMod "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	"google.golang.org/grpc/codes"
)

// InventoryReservationGet gets reservation details
func (c *Controller) InventoryReservationGet(ctx context.Context, req *pb.InventoryReservationGetRequest) (*pb.InventoryReservationGetResponse, error) {
	path := "inventory.controller.InventoryReservationGet"
	errBuilder := func(e *models.AppError) (*pb.InventoryReservationGetResponse, error) {
		return &pb.InventoryReservationGetResponse{Response: &pb.InventoryReservationGetResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}

	modelsCtx, ctxErr := models.ContextGet(ctx)
	if ctxErr != nil {
		return errBuilder(ctxErr)
	}
	internalErr := func(err error, details string) (*pb.InventoryReservationGetResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}))
	}
	sucBuilder := func(data *pb.InventoryReservationGetResponseData) (*pb.InventoryReservationGetResponse, error) {
		return &pb.InventoryReservationGetResponse{Response: &pb.InventoryReservationGetResponse_Data{Data: data}}, nil
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	// Get reservation
	reservation, err := c.store.InventoryReservationGetByToken(modelsCtx, nil, req.GetReservationToken())
	if err != nil {
		if err.ErrType == models.DBErrorTypeNoRows {
			return errBuilder(models.NewAppError(modelsCtx, path, "inventory.reservation.not_found", nil, err.Details, int(codes.NotFound), &models.AppErrorErrorsArgs{Err: err}))
		} else {
			return internalErr(err, "failed to get reservation")
		}
	}

	// Get reservation items
	reservationItems, err := c.store.InventoryReservationItemsGetByReservationID(modelsCtx, nil, reservation.Id)
	if err != nil {
		return internalErr(err, "failed to get reservation items")
	}

	ids := make([]string, 0, len(reservationItems))
	for _, item := range reservationItems {
		ids = append(ids, item.Id)
	}

	inventoryItems, err := c.store.InventoryItemGetByIDs(modelsCtx, ids)
	if err != nil {
		return internalErr(err, "failed to get inventory item")
	}

	// Convert to response format
	items := make([]*pb.InventoryReservationListItem, 0, len(reservationItems))
	for _, reservationItem := range inventoryItems {
		item, _ := utils.Find(reservationItems, func(i *pb.InventoryReservationItem) bool { return i.InventoryItemId == reservationItem.Id })

		// TODO: populate the status field
		items = append(items, &pb.InventoryReservationListItem{
			ProductId:         reservationItem.ProductId,
			VariantId:         reservationItem.VariantId,
			Sku:               reservationItem.Sku,
			QuantityRequested: uint32(item.Quantity),
			QuantityReserved:  uint32(item.Quantity),
		})
	}

	return sucBuilder(&pb.InventoryReservationGetResponseData{
		ReservationToken: reservation.ReservationToken,
		OrderId:          reservation.OrderId,
		Status:           intMod.GetInventoryReservationStatusFromString(reservation.Status),
		ExpiresAt:        reservation.ExpiresAt,
		Items:            items,
	})
}
