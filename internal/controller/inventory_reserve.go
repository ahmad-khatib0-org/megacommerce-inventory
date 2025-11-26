package controller

import (
	"context"
	"fmt"
	"time"

	intModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
)

// InventoryReserve reserves inventory for an order
func (c *Controller) InventoryReserve(ctx context.Context, req *pb.InventoryReserveRequest) (*pb.InventoryReserveResponse, error) {
	path := "inventory.controller.InventoryReserve"
	modelsCtx, ctxErr := models.ContextGet(ctx)

	errBuilder := func(e *models.AppError, tx pgx.Tx) (*pb.InventoryReserveResponse, error) {
		if tx != nil {
			rbErr := tx.Rollback(modelsCtx.Context)
			c.log.Errorf("%s: an error rolling back a transaction, err: %s", path, rbErr.Error())
		}
		return &pb.InventoryReserveResponse{Response: &pb.InventoryReserveResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}
	if ctxErr != nil {
		return errBuilder(ctxErr, nil)
	}
	internalErr := func(err error, details string, tx pgx.Tx) (*pb.InventoryReserveResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}), tx)
	}
	sucBuilder := func(data *pb.InventoryReserveResponseData) (*pb.InventoryReserveResponse, error) {
		return &pb.InventoryReserveResponse{Response: &pb.InventoryReserveResponse_Data{Data: data}}, nil
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	ar := models.AuditRecordNew(modelsCtx, intModels.EventNameInventoryReserve, models.EventStatusFail)
	defer func() {
		ar.AuditEventDataPriorState(intModels.InventoryReserveRequestAuditable(req))
		c.ProcessAudit(ar)
	}()

	tx, err := c.store.GetTx(modelsCtx.Context, pgx.TxOptions{})
	if err != nil {
		return internalErr(err, "failed to begin transaction", nil)
	}

	ttlSeconds := req.GetTtlSeconds()
	if ttlSeconds == 0 {
		ttlSeconds = 60
	}
	expiresAt := utils.TimeGetMillisFromTime(time.Now().Add(time.Duration(ttlSeconds) * time.Second))

	// Create reservation record
	reservationID := utils.NewID()
	reservationToken := "res_" + utils.NewID()
	err = c.store.InventoryReservationCreate(modelsCtx, tx, &pb.InventoryReservation{
		Id:               reservationID,
		ReservationToken: reservationToken,
		OrderId:          req.GetOrderId(),
		Status:           intModels.GetInventoryReservationStatus(pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RESERVED),
		ExpiresAt:        expiresAt,
		CreatedAt:        utils.TimeGetMillis(),
	})
	if err != nil {
		return internalErr(err, "failed to create reservation", tx)
	}

	partiallyErr := func(proID, varID string, quantity uint32) (*pb.InventoryReserveResponse, error) {
		key := fmt.Sprintf("%s.%s", proID, varID)
		ei := map[string]*models.AppErrorError{
			key: {ID: "orders.items.only_some_available", Params: map[string]any{"Quantity": quantity}},
		}
		errors := models.AppErrorErrorsArgs{Err: err, ErrorsInternal: ei}
		ai := models.NewAppError(ctxErr.Ctx, path, "orders.items.partially_available", nil, "", int(codes.Aborted), &errors)
		return errBuilder(ai, tx)
	}

	// Process each item
	reservationItems := make([]*pb.InventoryReservationListItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		// Check inventory availability
		inventory, errDB := c.store.InventoryItemGetByProductVariant(modelsCtx, tx, item.GetProductId(), item.GetVariantId())
		if errDB != nil {
			if errDB.ErrType == models.DBErrorTypeNoRows {
				key := fmt.Sprintf("%s.%s", item.GetProductId(), item.GetVariantId())
				errors := models.AppErrorErrorsArgs{
					Err: errDB,
					ErrorsInternal: map[string]*models.AppErrorError{
						key: {ID: "orders.items.not_found_in_inventory"},
					},
				}
				ai := models.NewAppError(ctxErr.Ctx, path, "error.not_found", nil, "", int(codes.NotFound), &errors)
				return errBuilder(ai, tx)
			}
		} else {
			return internalErr(errDB, "failed to query inventory_items table", tx)
		}

		if inventory.QuantityAvailable == 0 {
			key := fmt.Sprintf("%s.%s", item.GetProductId(), item.GetVariantId())
			ei := map[string]*models.AppErrorError{key: {ID: "orders.items.out_of_stock_for_variant"}}
			errors := models.AppErrorErrorsArgs{ErrorsInternal: ei}
			ai := models.NewAppError(ctxErr.Ctx, path, "orders.items.out_of_stock", nil, "", int(codes.Aborted), &errors)
			return errBuilder(ai, tx)
		}

		if inventory.QuantityAvailable < int32(item.GetQuantity()) {
			return partiallyErr(item.GetProductId(), item.GetVariantId(), uint32(inventory.QuantityAvailable))
		}

		// Reserve the inventory
		reserved, errDB := c.store.InventoryItemReserve(modelsCtx, tx, inventory.Id, int(item.GetQuantity()))
		if errDB != nil {
			return internalErr(errDB, "failed to reserve inventory", tx)
		}
		if !reserved {
			return partiallyErr(item.GetProductId(), item.GetVariantId(), uint32(inventory.QuantityAvailable))
		}

		// Create reservation item record
		errDB = c.store.InventoryReservationItemCreate(modelsCtx, tx, &pb.InventoryReservationItem{
			Id:              utils.NewID(),
			ReservationId:   reservationID,
			InventoryItemId: inventory.Id,
			Quantity:        int32(item.GetQuantity()),
			CreatedAt:       utils.TimeGetMillis(),
		})
		if errDB != nil {
			return internalErr(errDB, "failed to create a reservation item", tx)
		}

		// Add successfully reserved item to response
		reservationItems = append(reservationItems, &pb.InventoryReservationListItem{
			ProductId:         item.GetProductId(),
			VariantId:         item.GetVariantId(),
			Sku:               item.GetSku(),
			QuantityRequested: item.GetQuantity(),
			QuantityReserved:  item.GetQuantity(),
			Status:            pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_RESERVED,
		})
	}

	// Update reservation status
	resStatus := pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RESERVED
	err = c.store.InventoryReservationUpdateStatus(modelsCtx, tx, reservationID, intModels.GetInventoryReservationStatus(resStatus))
	if err != nil {
		return internalErr(err, "failed to update reservation status", tx)
	}

	// Commit transaction
	if err := tx.Commit(modelsCtx.Context); err != nil {
		return internalErr(err, "failed to commit the overall reservation transaction", tx)
	}

	ar.Success()

	return sucBuilder(&pb.InventoryReserveResponseData{
		ReservationToken: reservationToken,
		Status:           resStatus,
		Items:            reservationItems,
	})
}
