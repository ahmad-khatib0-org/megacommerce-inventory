package dbstore

import (
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	"github.com/jackc/pgx/v5"
)

// InventoryReservationGetByToken gets a reservation by its token
func (is *InventoryStore) InventoryReservationGetByToken(ctx *models.Context, tx pgx.Tx, token string) (*pb.InventoryReservation, *models.DBError) {
	stmt := `
		SELECT 
			id, 
			reservation_token, 
			order_id, 
			status, 
			expires_at, 
			created_at, 
			updated_at
		FROM inventory_reservations 
		WHERE reservation_token = $1;
  `

	var ir pb.InventoryReservation
	var updatedAt int64
	err := tx.QueryRow(ctx.Ctx(), stmt, token).Scan(
		&ir.Id,
		&ir.ReservationToken,
		&ir.OrderId,
		&ir.Status,
		&ir.ExpiresAt,
		&ir.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryReservationGetByToken", tx)
	}

	if updatedAt > 0 {
		ir.UpdatedAt = &updatedAt
	}

	return &ir, nil
}

// InventoryReservationCreate creates a new reservation
func (is *InventoryStore) InventoryReservationCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryReservation) *models.DBError {
	stmt := `
		INSERT INTO inventory_reservations (
			id,
			reservation_token,
			order_id,
			status,
			expires_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		params.Id,
		params.ReservationToken,
		params.OrderId,
		params.Status,
		params.ExpiresAt,
		params.CreatedAt,
		params.UpdatedAt,
	)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryReservationCreate", tx)
}

// InventoryReservationUpdateStatus updates the status of a reservation
func (is *InventoryStore) InventoryReservationUpdateStatus(ctx *models.Context, tx pgx.Tx, id string, status string) *models.DBError {
	stmt := `
		UPDATE inventory_reservations 
		SET status = $1, updated_at = $2
		WHERE id = $3
  `

	_, err := tx.Exec(ctx.Ctx(), stmt, status, utils.TimeGetMillis(), id)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryReservationUpdateStatus", tx)
}
