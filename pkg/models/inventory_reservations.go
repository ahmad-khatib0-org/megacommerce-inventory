package models

import (
	"strings"
	"time"

	pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"
)

func GetInventoryReservationStatus(status pb.InventoryReservationStatus) string {
	switch status {
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RESERVED:
		return "RESERVED"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_PARTIALLY_RESERVED:
		return "PARTIALLY_RESERVED"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_NOT_RESERVED:
		return "NOT_RESERVED"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_PENDING:
		return "PENDING"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RELEASED:
		return "RELEASED"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_FULFILLED:
		return "FULFILLED"
	case pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_UNSPECIFIED:
		fallthrough
	default:
		return "UNSPECIFIED"
	}
}

func GetInventoryReservationStatusFromString(statusStr string) pb.InventoryReservationStatus {
	switch strings.ToUpper(statusStr) {
	case "RESERVED":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RESERVED
	case "PARTIALLY_RESERVED":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_PARTIALLY_RESERVED
	case "NOT_RESERVED":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_NOT_RESERVED
	case "PENDING":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_PENDING
	case "RELEASED":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_RELEASED
	case "FULFILLED":
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_FULFILLED
	default:
		return pb.InventoryReservationStatus_INVENTORY_RESERVATION_STATUS_UNSPECIFIED
	}
}

func GetInventoryReservationItemStatus(status pb.InventoryReservationItemStatus) string {
	switch status {
	case pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_RESERVED:
		return "RESERVED"
	case pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_NOT_RESERVED:
		return "NOT_RESERVED"
	case pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_OUT_OF_STOCK:
		return "OUT_OF_STOCK"
	case pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_UNSPECIFIED:
		fallthrough
	default:
		return "UNSPECIFIED"
	}
}

func GetInventoryReservationItemStatusFromString(statusStr string) pb.InventoryReservationItemStatus {
	switch strings.ToUpper(statusStr) {
	case "RESERVED":
		return pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_RESERVED
	case "NOT_RESERVED":
		return pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_NOT_RESERVED
	case "OUT_OF_STOCK":
		return pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_OUT_OF_STOCK
	default:
		return pb.InventoryReservationItemStatus_INVENTORY_RESERVATION_ITEM_STATUS_UNSPECIFIED
	}
}

func GetInventoryUpdateOperation(operation pb.InventoryUpdateOperation) string {
	switch operation {
	case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_SET:
		return "SET"
	case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_ADD:
		return "ADD"
	case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_SUBTRACT:
		return "SUBTRACT"
	case pb.InventoryUpdateOperation_INVENTORY_UPDATE_OPERATION_UNSPECIFIED:
		fallthrough
	default:
		return "UNSPECIFIED"
	}
}

func InventoryReserveRequestAuditable(req *pb.InventoryReserveRequest) map[string]any {
	if req == nil {
		return map[string]any{}
	}

	items := make([]map[string]any, len(req.Items))
	for i, item := range req.Items {
		items[i] = map[string]any{
			"product_id": item.ProductId,
			"variant_id": item.VariantId,
			"sku":        item.Sku,
			"quantity":   item.Quantity,
		}
	}

	return map[string]any{
		"order_id":    req.OrderId,
		"items":       items,
		"ttl_seconds": req.TtlSeconds,
		"expires_at":  time.Now().Add(time.Duration(req.TtlSeconds) * time.Second).Unix(),
	}
}

func InventoryUpdateRequestAuditable(req *pb.InventoryUpdateRequest) map[string]any {
	if req == nil {
		return map[string]any{}
	}

	items := make([]map[string]any, len(req.Items))
	for i, item := range req.Items {
		items[i] = map[string]any{
			"product_id": item.ProductId,
			"variant_id": item.VariantId,
			"sku":        item.Sku,
			"operation":  GetInventoryUpdateOperation(item.Operation),
			"quantity":   item.Quantity,
		}
	}

	return map[string]any{
		"items":  items,
		"reason": req.Reason,
	}
}
