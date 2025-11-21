package dbstore

import (
	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/utils"
	"github.com/jackc/pgx/v5"
)

// InventoryItemGetByProductVariant gets an inventory item by product and variant
func (is *InventoryStore) InventoryItemGetByProductVariant(ctx *models.Context, tx pgx.Tx, productID string, variantID string) (*pb.InventoryItem, *models.DBError) {
	stmt := `
		SELECT 
			id, 
			product_id, 
			variant_id, 
			sku, 
			quantity_available, 
			quantity_reserved, 
			quantity_total, 
			location_id, 
			metadata, 
			created_at, 
			updated_at
		FROM inventory_items 
		WHERE product_id = $1 AND variant_id = $2
		FOR UPDATE;
  `

	var ii pb.InventoryItem
	var updatedAt int64
	err := tx.QueryRow(ctx.Ctx(), stmt, productID, variantID).Scan(
		&ii.Id,
		&ii.ProductId,
		&ii.VariantId,
		&ii.Sku,
		&ii.QuantityAvailable,
		&ii.QuantityReserved,
		&ii.QuantityTotal,
		&ii.LocationId,
		&ii.Metadata,
		&ii.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		return nil, models.HandleDBError(ctx, err, "inventory.store.InventoryItemGetByProductVariant", tx)
	}

	if updatedAt > 0 {
		ii.UpdatedAt = &updatedAt
	}

	return &ii, nil
}

// InventoryItemCreate creates a new inventory item
func (is *InventoryStore) InventoryItemCreate(ctx *models.Context, tx pgx.Tx, params *pb.InventoryItem) *models.DBError {
	stmt := `
		INSERT INTO inventory_items (
			id, 
			product_id, 
			variant_id, 
			sku, 
			quantity_available, 
			quantity_reserved, 
			quantity_total, 
			location_id, 
			metadata, 
			created_at, 
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
  `

	_, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		params.Id,
		params.ProductId,
		params.VariantId,
		params.Sku,
		params.QuantityAvailable,
		params.QuantityReserved,
		params.QuantityTotal,
		params.LocationId,
		params.Metadata,
		params.CreatedAt,
		params.UpdatedAt,
	)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryItemCreate", tx)
}

// InventoryItemReserve reserves inventory for an item, and returns sufficient = false
// if the requested quantity can't be reserved (quantity_available is not enough)
func (is *InventoryStore) InventoryItemReserve(ctx *models.Context, tx pgx.Tx, id string, quantity int) (bool, *models.DBError) {
	stmt := `
			UPDATE inventory_items 
			SET 
					quantity_reserved = quantity_reserved + $1,
					quantity_available = quantity_available - $1,
					updated_at = $2
			WHERE id = $3 AND quantity_available >= $1
    `

	result, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		quantity,
		utils.TimeGetMillis(),
		id,
	)
	if err != nil {
		return false, models.HandleDBError(ctx, err, "inventory.store.InventoryItemReserve", tx)
	}

	if result.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

// InventoryItemRelease releases reserved inventory for an item, it returns
//
// release = false if the query didn't update this row, which means that the
//
// quantity we are about to release is bigger than the current quantity_reserved value!
func (is *InventoryStore) InventoryItemRelease(ctx *models.Context, tx pgx.Tx, id string, quantity int32) (bool, *models.DBError) {
	// AND quantity_reserved >= $1 to prevent making quantity_reserved less than 0
	stmt := `
			UPDATE inventory_items 
			SET 
				quantity_reserved = quantity_reserved - $1,
				quantity_available = quantity_available + $1,
				updated_at = $2
			WHERE id = $3 AND quantity_reserved >= $1
    `

	result, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		quantity,
		utils.TimeGetMillis(),
		id,
	)
	if err != nil {
		return false, models.HandleDBError(ctx, err, "inventory.store.InventoryItemRelease", tx)
	}

	if result.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

// InventoryItemUpdate updates an inventory item
func (is *InventoryStore) InventoryItemUpdate(ctx *models.Context, tx pgx.Tx, id string, quantityTotal int, quantityReserved int32, quantityAvailable int) *models.DBError {
	stmt := `
		UPDATE inventory_items 
		SET 
			quantity_total = $1,
			quantity_reserved = $2,
			quantity_available = $3,
			updated_at = $4
		WHERE id = $5
  `

	_, err := tx.Exec(
		ctx.Ctx(),
		stmt,
		quantityTotal,
		quantityReserved,
		quantityAvailable,
		utils.TimeGetMillis(),
		id,
	)

	return models.HandleDBError(ctx, err, "inventory.store.InventoryItemUpdate", tx)
}
