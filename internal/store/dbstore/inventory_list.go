package dbstore

import (
	"fmt"

	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
)

// InventoryList gets a paginated list of inventory items for a supplier
func (is *InventoryStore) InventoryList(ctx *models.Context, userID string, limit int32, lastID string) ([]*pb.InventoryListItem, *models.DBError) {
	whereClause := ""
	args := []any{userID}

	if lastID != "" {
		whereClause = " AND ii.id < $2 "
		args = append(args, lastID)
	}

	limitParam := "$2"
	if lastID != "" {
		limitParam = "$3"
	}

	stmt := fmt.Sprintf(`
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
				WHEN ii.quantity_available < 10 THEN 'low_stock'
				ELSE 'in_stock'
			END as status,
			ii.created_at,
			ii.updated_at
		FROM inventory_items ii
		JOIN products p ON ii.product_id = p.id
		WHERE p.user_id = $1 %s
		ORDER BY ii.id DESC
		LIMIT %s
	`, whereClause, limitParam)

	args = append(args, limit)

	rows, err := is.db.Query(ctx.Ctx(), stmt, args...)
	if err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryList ", nil)
	}
	defer rows.Close()

	result := make([]*pb.InventoryListItem, 0)
	for rows.Next() {
		var item pb.InventoryListItem
		var status string

		err = rows.Scan(
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
			return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryList", nil)
		}

		item.Status = status
		result = append(result, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryList", nil)
	}

	return result, nil
}
