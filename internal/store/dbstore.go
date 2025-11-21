// Package store contains the interfaces that must be implemented for handling db queries
package store

import (
	"context"

	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryDBStore interface {
	GetTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	InventoryReservationGetByToken(ctx *models.Context, tx *pgxpool.Tx, token string) (*pb.InventoryReservation, *models.DBError)
}
