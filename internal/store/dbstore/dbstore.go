// Package dbstore store/update/delete/retrieve data in the database
package dbstore

import (
	"context"

	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/store"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryStore struct {
	db *pgxpool.Pool
}

func (is *InventoryStore) GetTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, *models.DBError) {
	tx, err := is.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, &models.DBError{ErrType: models.DBErrorTypeStartTransaction, Err: err, Msg: "failed to start a db transaction"}
	}
	return tx, nil
}

func NewInventoryStore(pool *pgxpool.Pool) store.InventoryDBStore {
	return &InventoryStore{db: pool}
}
