package models

import (
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

