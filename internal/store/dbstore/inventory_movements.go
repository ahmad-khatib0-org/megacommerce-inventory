package dbstore

import (
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5"
)

// InventoryMovementCreate creates a new inventory movement
func (is *InventoryStore) InventoryMovementCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryMovement) *models.DBError {
	stmt := `
		INSERT INTO inventory_movements (
			id,
			inventory_item_id,
			movement_type,
			quantity,
			reference_id,
			reason,
			metadata,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  `

	_, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		params.Id,
		params.InventoryItemId,
		params.MovementType,
		params.Quantity,
		params.ReferenceId,
		params.Reason,
		params.Metadata,
		params.CreatedAt,
	)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryMovementCreate", tx)
}
