// Package store contains the interfaces that must be implemented for handling db queries
package store

import (
	"context"

	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
)

type InventoryDBStore interface {
	GetTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	InventoryReservationGetByToken(ctx *models.Context, tx pgx.Tx, token string) (*pb.InventoryReservation, *models.DBError)
	InventoryReservationCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryReservation) *models.DBError
	InventoryReservationUpdateStatus(ctx *models.Context, tx pgx.Tx, id string, status string) *models.DBError
	InventoryItemGetByProductVariant(ctx *models.Context, tx pgx.Tx, productID string, variantID string) (*pb.InventoryItem, *models.DBError)
	InventoryItemCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryItem) *models.DBError
	// InventoryItemReserve reserves inventory for an item, and returns sufficient = false
	//
	// if the requested quantity can't be reserved (quantity_available is not enough)
	InventoryItemReserve(ctx *models.Context, tx pgx.Tx, id string, quantity int) (bool, *models.DBError)
	// InventoryItemRelease releases reserved inventory for an item, it returns
	//
	// release = false if the query didn't update this row, which means that the
	//
	// quantity we are about to release is bigger than the current quantity_reserved value!
	InventoryItemRelease(ctx *models.Context, tx pgx.Tx, id string, quantity int32) (bool, *models.DBError)
	// InventoryReservationItemCreate creates a new reservation item
	InventoryReservationItemCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryReservationItem) *models.DBError
}
