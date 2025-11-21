package models

import pb "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/inventory/v1"

func GetInventoryMovementType(movementType pb.InventoryMovementType) string {
	switch movementType {
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_IN:
		return "IN"
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_OUT:
		return "OUT"
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_ADJUSTMENT:
		return "ADJUSTMENT"
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_RESERVATION:
		return "RESERVATION"
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_RELEASE:
		return "RELEASE"
	case pb.InventoryMovementType_INVENTORY_MOVEMENT_TYPE_UNSPECIFIED:
		fallthrough
	default:
		return "UNSPECIFIED"
	}
}
