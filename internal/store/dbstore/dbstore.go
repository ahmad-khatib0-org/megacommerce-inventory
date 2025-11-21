// Package dbstore store/update/delete/retrieve data in the database
package dbstore

import (
	"context"

	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryStore struct {
	db *pgxpool.Pool
}

func (is *InventoryStore) GetTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return is.db.BeginTx(ctx, opts)
}

func NewInventoryStore(pool *pgxpool.Pool) store.InventoryDBStore {
	return &InventoryStore{db: pool}
}
