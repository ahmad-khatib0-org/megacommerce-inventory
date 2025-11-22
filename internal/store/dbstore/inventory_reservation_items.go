package dbstore

import (
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
)

// InventoryReservationItemCreate creates a new reservation item
func (is *InventoryStore) InventoryReservationItemCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryReservationItem) *models.DBError {
	stmt := `
		INSERT INTO inventory_reservation_items (
			id, 
			reservation_id, 
			inventory_item_id, 
			quantity, 
			created_at
		) VALUES ($1, $2, $3, $4, $5)
    `

	_, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		params.Id,
		params.ReservationId,
		params.InventoryItemId,
		params.Quantity,
		params.CreatedAt,
	)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryReservationItemCreate", tx)
}

// InventoryReservationItemsGetByReservationID gets all items for a reservation
func (is *InventoryStore) InventoryReservationItemsGetByReservationID(ctx *models.Context, tx pgx.Tx, reservationID string) ([]*pb.InventoryReservationItem, *models.DBError) {
	stmt := `
		SELECT 
			id,
			reservation_id,
			inventory_item_id,
			quantity,
			created_at
		FROM inventory_reservation_items
		WHERE reservation_id = $1
  `

	var rows pgx.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(ctx.Context, stmt, reservationID)
	} else {
		rows, err = is.db.Query(ctx.Context, stmt, reservationID)
	}
	if err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryReservationItemsGetByReservationID", tx)
	}
	defer rows.Close()

	var items []*pb.InventoryReservationItem
	for rows.Next() {
		var item pb.InventoryReservationItem
		err := rows.Scan(
			&item.Id,
			&item.ReservationId,
			&item.InventoryItemId,
			&item.Quantity,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryReservationItemsGetByReservationID", tx)
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryReservationItemsGetByReservationID", tx)
	}

	return items, nil
}
