package controller

import (
	"context"
	"time"

	intModels "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	pbSh "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/shared/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
)

// InventoryUpdate updates inventory levels
func (c *Controller) InventoryUpdate(ctx context.Context, req *pb.InventoryUpdateRequest) (*pb.InventoryUpdateResponse, error) {
	path := "inventory.controller.InventoryUpdate"
	modelsCtx, ctxErr := models.ContextGet(ctx)
	errBuilder := func(e *models.AppError, tx pgx.Tx) (*pb.InventoryUpdateResponse, error) {
		if tx != nil {
			rbErr := tx.Rollback(modelsCtx.Context)
			c.log.Errorf("%s: an error rolling back a transaction, err: %s", path, rbErr.Error())
		}
		return &pb.InventoryUpdateResponse{Response: &pb.InventoryUpdateResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}

	if ctxErr != nil {
		return errBuilder(ctxErr, nil)
	}
	internalErr := func(err error, details string, tx pgx.Tx) (*pb.InventoryUpdateResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}), tx)
	}
	sucBuilder := func(data *pbSh.SuccessResponseData) (*pb.InventoryUpdateResponse, error) {
		return &pb.InventoryUpdateResponse{Response: &pb.InventoryUpdateResponse_Data{Data: data}}, nil
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	ar := models.AuditRecordNew(modelsCtx, intModels.EventNameInventoryUpdate, models.EventStatusFail)
	defer func() {
		ar.AuditEventDataPriorState(intModels.InventoryUpdateRequestAuditable(req))
		c.ProcessAudit(ar)
	}()

	// if err := models.InventoryUpdateRequestIsValid(modelsCtx, req); err != nil {
	// 	return errBuilder(err)
	// }

	tx, err := c.store.GetTx(modelsCtx.Context, pgx.TxOptions{})
	if err != nil {
		return internalErr(err, err.Msg, tx)
	}

	// Process each item
	for _, item := range req.GetItems() {
		inventory, err := c.store.InventoryItemGetByProductVariant(modelsCtx, tx, item.GetProductId(), item.GetVariantId())
		if err != nil {
			if err.ErrType == models.DBErrorTypeNoRows {
				errors := models.AppErrorErrorsArgs{
					Err: err,
					ErrorsInternal: map[string]*models.AppErrorError{
						item.GetVariantId(): {ID: "orders.items.not_found_in_inventory"},
					},
				}
				ai := models.NewAppError(modelsCtx, path, "error.not_found", nil, "", int(codes.NotFound), &errors)
				return errBuilder(ai, tx)
			} else {
				return internalErr(err, "failed to update inventory", tx)
			}
		}
		var newQuantityTotal int
		var newQuantityAvailable int
		var movementType string

		// Update inventory based on operation
		switch item.GetOperation() {
		case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_SET:
			newQuantityTotal = int(item.GetQuantity())
			newQuantityAvailable = newQuantityTotal - int(inventory.QuantityReserved)
			movementType = intModels.GetInventoryMovementType(pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_ADJUSTMENT)
		case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_ADD:
			newQuantityTotal = int(inventory.QuantityTotal + int32(item.GetQuantity()))
			newQuantityAvailable = int(inventory.QuantityAvailable + int32(item.GetQuantity()))
			movementType = intModels.GetInventoryMovementType(pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_IN)
		case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_SUBTRACT:
			newQuantityTotal = int(inventory.QuantityTotal - int32(item.GetQuantity()))
			newQuantityAvailable = int(inventory.QuantityAvailable - int32(item.GetQuantity()))
			movementType = intModels.GetInventoryMovementType(pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_OUT)
		default:
			return errBuilder(models.NewAppError(modelsCtx, path, "inventory.update.invalid_operation", nil, "", int(codes.InvalidArgument), nil), tx)
		}

		// Check if we have enough available inventory for subtraction
		if item.GetOperation() == pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_SUBTRACT &&
			int32(item.GetQuantity()) > inventory.QuantityAvailable {
			id := "inventory.update.insufficient_available"
			errors := models.AppErrorErrorsArgs{
				ErrorsInternal: map[string]*models.AppErrorError{item.GetVariantId(): {ID: id}},
			}
			return errBuilder(models.NewAppError(modelsCtx, path, id, nil, "", int(codes.InvalidArgument), &errors), tx)
		}

		err = c.store.InventoryItemUpdate(modelsCtx, tx, inventory.Id, newQuantityTotal, inventory.QuantityReserved, newQuantityAvailable)
		if err != nil {
			return internalErr(err, "failed to update inventory", tx)
		}

		err = c.store.InventoryMovementCreate(modelsCtx, tx, &pb.InventoryMovement{
			Id:              utils.NewID(),
			InventoryItemId: inventory.Id,
			MovementType:    movementType,
			Quantity:        int32(item.GetQuantity()),
			Reason:          req.Reason,
			CreatedAt:       utils.TimeGetMillis(),
		})
		if err != nil {
			return internalErr(err, "failed to create inventory movement", tx)
		}
	}

	if err := tx.Commit(modelsCtx.Context); err != nil {
		return internalErr(err, "failed to commit transaction", tx)
	}

	ar.Success()

	msg := models.Tr(modelsCtx.AcceptLanguage, "inventory.update.success", nil)
	return sucBuilder(&pbSh.SuccessResponseData{Message: &msg})
}
