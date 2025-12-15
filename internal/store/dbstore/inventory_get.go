package dbstore

import (
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
)

// InventoryItemGetByID gets a single inventory item by ID with user validation
func (is *InventoryStore) InventoryItemGetByID(ctx *models.Context, userID string, itemID string) (*pb.InventoryListItem, *models.DBError) {
	stmt := `
		SELECT 
			ii.id,
			ii.product_id,
			ii.variant_id,
			ii.sku,
			p.title,
			ii.quantity_total,
			ii.quantity_available,
			ii.quantity_reserved,
			CASE 
				WHEN ii.quantity_available = 0 THEN 'out_of_stock'
				WHEN ii.quantity_available <= (ii.quantity_total * 0.1) THEN 'low_stock'
				ELSE 'in_stock'
			END as status,
			ii.created_at,
			ii.updated_at
		FROM inventory_items ii
		JOIN products p ON ii.product_id = p.id
		WHERE ii.id = $1 AND p.user_id = $2
	`

	var item pb.InventoryListItem
	var status string

	err := is.db.QueryRow(ctx.Ctx(), stmt, itemID, userID).Scan(
		&item.Id,
		&item.ProductId,
		&item.VariantId,
		&item.Sku,
		&item.ProductName,
		&item.QuantityTotal,
		&item.QuantityAvailable,
		&item.QuantityReserved,
		&status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryItemGetByID", nil)
	}

	item.Status = status
	return &item, nil
}
