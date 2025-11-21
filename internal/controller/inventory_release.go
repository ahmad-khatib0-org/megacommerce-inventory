package controller

import (
	"context"
	"fmt"
	"time"

	modelsInt "github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	pbSh "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/shared/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
)

// InventoryRelease releases inventory reservation
func (c *Controller) InventoryRelease(ctx context.Context, req *pb.InventoryReleaseRequest) (*pb.InventoryReleaseResponse, error) {
	path := "inventory.controller.InventoryRelease"
	modelsCtx, ctxErr := models.ContextGet(ctx)
	errBuilder := func(e *models.AppError, tx pgx.Tx) (*pb.InventoryReleaseResponse, error) {
		if tx != nil {
			rbErr := tx.Rollback(modelsCtx.Context)
			c.log.Errorf("%s: an error rolling back a transaction, err: %s", path, rbErr.Error())
		}
		return &pb.InventoryReleaseResponse{Response: &pb.InventoryReleaseResponse_Error{Error: models.AppErrorToProto(e)}}, nil
	}

	if ctxErr != nil {
		return errBuilder(ctxErr, nil)
	}
	internalErr := func(err error, details string, tx pgx.Tx) (*pb.InventoryReleaseResponse, error) {
		return errBuilder(models.NewAppError(modelsCtx, path, models.ErrMsgInternal, nil, details, int(codes.Internal), &models.AppErrorErrorsArgs{Err: err}), tx)
	}
	sucBuilder := func(data *pbSh.SuccessResponseData) (*pb.InventoryReleaseResponse, error) {
		return &pb.InventoryReleaseResponse{Response: &pb.InventoryReleaseResponse_Data{Data: data}}, nil
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	modelsCtx.Context = rctx

	ar := models.AuditRecordNew(modelsCtx, modelsInt.EventNameInventoryRelease, models.EventStatusFail)
	defer c.ProcessAudit(ar)

	tx, err := c.store.GetTx(modelsCtx.Context, pgx.TxOptions{})
	if err != nil {
		return internalErr(err, "failed to begin transaction", tx)
	}

	// Get reservation
	reservation, err := c.store.InventoryReservationGetByToken(modelsCtx, tx, req.GetReservationToken())
	if err != nil {
		if err.ErrType == models.DBErrorTypeNoRows {
			return errBuilder(models.NewAppError(modelsCtx, path, "inventory.reservation.not_found", nil, err.Details, int(codes.NotFound), &models.AppErrorErrorsArgs{Err: err}), tx)
		} else {
			return internalErr(err, "failed to get reservation", tx)
		}
	}

	// Check if reservation is already released or fulfilled
	if reservation.Status == pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RELEASED.String() ||
		reservation.Status == pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_FULFILLED.String() {
		return errBuilder(models.NewAppError(modelsCtx, path, "inventory.reservation.already_processed", nil, "", int(codes.InvalidArgument), nil), tx)
	}

	// Get reservation items
	reservationItems, err := c.store.InventoryReservationItemsGetByReservationID(modelsCtx, tx, reservation.Id)
	if err != nil {
		return internalErr(err, "failed to get reservation items", tx)
	}

	// Release inventory for each item
	for _, reservationItem := range reservationItems {
		released, errDB := c.store.InventoryItemRelease(modelsCtx, tx, reservationItem.InventoryItemId, reservationItem.Quantity)
		if errDB != nil {
			return internalErr(err, "failed to release inventory", tx)
		}
		if !released {
			// TODO: this should not happen, and should be added to DLQ to be reviewed
			msg := "The requested quantity to be released is bigger than the quantity_reserved value"
			return internalErr(err, fmt.Sprintf("failed to release inventory, %s", msg), tx)
		}
	}

	// Update reservation status
	err = c.store.InventoryReservationUpdateStatus(modelsCtx, tx, reservation.Id, pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RELEASED.String())
	if err != nil {
		return internalErr(err, "failed to update reservation status", tx)
	}

	// Commit transaction
	if err := tx.Commit(modelsCtx.Context); err != nil {
		return internalErr(err, "failed to commit transaction", tx)
	}

	ar.Success()

	msg := models.Tr(modelsCtx.AcceptLanguage, "inventory.release.success", nil)
	return sucBuilder(&pbSh.SuccessResponseData{Message: &msg})
}
