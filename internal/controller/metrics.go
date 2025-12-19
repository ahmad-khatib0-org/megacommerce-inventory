package controller

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type MetricsCollector struct {
	// Inventory Get metrics
	inventoryGetTotal    metric.Int64Counter
	inventoryGetErrors   metric.Int64Counter
	inventoryGetDuration metric.Float64Histogram

	// Inventory List metrics
	inventoryListTotal    metric.Int64Counter
	inventoryListErrors   metric.Int64Counter
	inventoryListDuration metric.Float64Histogram

	// Inventory Reserve metrics
	inventoryReserveTotal       metric.Int64Counter
	inventoryReserveErrors      metric.Int64Counter
	inventoryReserveDuration    metric.Float64Histogram
	inventoryItemsReserved      metric.Int64Counter
	inventoryReservationsFailed metric.Int64Counter

	// Inventory Release metrics
	inventoryReleaseTotal    metric.Int64Counter
	inventoryReleaseErrors   metric.Int64Counter
	inventoryReleaseDuration metric.Float64Histogram
	inventoryItemsReleased   metric.Int64Counter

	// Inventory Update metrics
	inventoryUpdateTotal    metric.Int64Counter
	inventoryUpdateErrors   metric.Int64Counter
	inventoryUpdateDuration metric.Float64Histogram

	// Inventory Reservation Get metrics
	inventoryReservationGetTotal    metric.Int64Counter
	inventoryReservationGetErrors   metric.Int64Counter
	inventoryReservationGetDuration metric.Float64Histogram

	// Transaction metrics
	transactionsStarted metric.Int64Counter
	transactionsSuccess metric.Int64Counter
	transactionsFailed  metric.Int64Counter
	transactionDuration metric.Float64Histogram

	// Database operation metrics
	dbOperationsTotal   metric.Int64Counter
	dbOperationErrors   metric.Int64Counter
	dbOperationDuration metric.Float64Histogram
}

func NewMetricsCollector() *MetricsCollector {
	meter := otel.GetMeterProvider().Meter("megacommerce-inventory", metric.WithInstrumentationVersion("0.1.0"))

	mc := &MetricsCollector{}

	// Inventory Get metrics
	mc.inventoryGetTotal, _ = meter.Int64Counter("inventory_get_total",
		metric.WithDescription("Total inventory get requests"))
	mc.inventoryGetErrors, _ = meter.Int64Counter("inventory_get_errors_total",
		metric.WithDescription("Total inventory get errors"))
	mc.inventoryGetDuration, _ = meter.Float64Histogram("inventory_get_duration_seconds",
		metric.WithDescription("Inventory get request duration in seconds"))

	// Inventory List metrics
	mc.inventoryListTotal, _ = meter.Int64Counter("inventory_list_total",
		metric.WithDescription("Total inventory list requests"))
	mc.inventoryListErrors, _ = meter.Int64Counter("inventory_list_errors_total",
		metric.WithDescription("Total inventory list errors"))
	mc.inventoryListDuration, _ = meter.Float64Histogram("inventory_list_duration_seconds",
		metric.WithDescription("Inventory list request duration in seconds"))

	// Inventory Reserve metrics
	mc.inventoryReserveTotal, _ = meter.Int64Counter("inventory_reserve_total",
		metric.WithDescription("Total inventory reserve requests"))
	mc.inventoryReserveErrors, _ = meter.Int64Counter("inventory_reserve_errors_total",
		metric.WithDescription("Total inventory reserve errors"))
	mc.inventoryReserveDuration, _ = meter.Float64Histogram("inventory_reserve_duration_seconds",
		metric.WithDescription("Inventory reserve request duration in seconds"))
	mc.inventoryItemsReserved, _ = meter.Int64Counter("inventory_items_reserved_total",
		metric.WithDescription("Total inventory items successfully reserved"))
	mc.inventoryReservationsFailed, _ = meter.Int64Counter("inventory_reservations_failed_total",
		metric.WithDescription("Total failed reservation attempts"))

	// Inventory Release metrics
	mc.inventoryReleaseTotal, _ = meter.Int64Counter("inventory_release_total",
		metric.WithDescription("Total inventory release requests"))
	mc.inventoryReleaseErrors, _ = meter.Int64Counter("inventory_release_errors_total",
		metric.WithDescription("Total inventory release errors"))
	mc.inventoryReleaseDuration, _ = meter.Float64Histogram("inventory_release_duration_seconds",
		metric.WithDescription("Inventory release request duration in seconds"))
	mc.inventoryItemsReleased, _ = meter.Int64Counter("inventory_items_released_total",
		metric.WithDescription("Total inventory items successfully released"))

	// Inventory Update metrics
	mc.inventoryUpdateTotal, _ = meter.Int64Counter("inventory_update_total",
		metric.WithDescription("Total inventory update requests"))
	mc.inventoryUpdateErrors, _ = meter.Int64Counter("inventory_update_errors_total",
		metric.WithDescription("Total inventory update errors"))
	mc.inventoryUpdateDuration, _ = meter.Float64Histogram("inventory_update_duration_seconds",
		metric.WithDescription("Inventory update request duration in seconds"))

	// Inventory Reservation Get metrics
	mc.inventoryReservationGetTotal, _ = meter.Int64Counter("inventory_reservation_get_total",
		metric.WithDescription("Total inventory reservation get requests"))
	mc.inventoryReservationGetErrors, _ = meter.Int64Counter("inventory_reservation_get_errors_total",
		metric.WithDescription("Total inventory reservation get errors"))
	mc.inventoryReservationGetDuration, _ = meter.Float64Histogram("inventory_reservation_get_duration_seconds",
		metric.WithDescription("Inventory reservation get request duration in seconds"))

	// Transaction metrics
	mc.transactionsStarted, _ = meter.Int64Counter("transactions_started_total",
		metric.WithDescription("Total transactions started"))
	mc.transactionsSuccess, _ = meter.Int64Counter("transactions_success_total",
		metric.WithDescription("Total successful transactions"))
	mc.transactionsFailed, _ = meter.Int64Counter("transactions_failed_total",
		metric.WithDescription("Total failed transactions"))
	mc.transactionDuration, _ = meter.Float64Histogram("transaction_duration_seconds",
		metric.WithDescription("Transaction duration in seconds"))

	// Database operation metrics
	mc.dbOperationsTotal, _ = meter.Int64Counter("db_operations_total",
		metric.WithDescription("Total database operations"))
	mc.dbOperationErrors, _ = meter.Int64Counter("db_operation_errors_total",
		metric.WithDescription("Total database operation errors"))
	mc.dbOperationDuration, _ = meter.Float64Histogram("db_operation_duration_seconds",
		metric.WithDescription("Database operation duration in seconds"))

	return mc
}

func (m *MetricsCollector) RecordInventoryGetRequest(success bool, duration float64) {
	ctx := context.Background()
	m.inventoryGetTotal.Add(ctx, 1)
	m.inventoryGetDuration.Record(ctx, duration)
	if !success {
		m.inventoryGetErrors.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordInventoryListRequest(success bool, duration float64) {
	ctx := context.Background()
	m.inventoryListTotal.Add(ctx, 1)
	m.inventoryListDuration.Record(ctx, duration)
	if !success {
		m.inventoryListErrors.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordInventoryReserveRequest(success bool, duration float64, itemsCount int64) {
	ctx := context.Background()
	m.inventoryReserveTotal.Add(ctx, 1)
	m.inventoryReserveDuration.Record(ctx, duration)
	if success {
		m.inventoryItemsReserved.Add(ctx, itemsCount)
	} else {
		m.inventoryReserveErrors.Add(ctx, 1)
		m.inventoryReservationsFailed.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordInventoryReleaseRequest(success bool, duration float64, itemsCount int64) {
	ctx := context.Background()
	m.inventoryReleaseTotal.Add(ctx, 1)
	m.inventoryReleaseDuration.Record(ctx, duration)
	if success {
		m.inventoryItemsReleased.Add(ctx, itemsCount)
	} else {
		m.inventoryReleaseErrors.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordInventoryUpdateRequest(success bool, duration float64) {
	ctx := context.Background()
	m.inventoryUpdateTotal.Add(ctx, 1)
	m.inventoryUpdateDuration.Record(ctx, duration)
	if !success {
		m.inventoryUpdateErrors.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordInventoryReservationGetRequest(success bool, duration float64) {
	ctx := context.Background()
	m.inventoryReservationGetTotal.Add(ctx, 1)
	m.inventoryReservationGetDuration.Record(ctx, duration)
	if !success {
		m.inventoryReservationGetErrors.Add(ctx, 1)
	}
}

func (m *MetricsCollector) RecordTransactionStarted() {
	ctx := context.Background()
	m.transactionsStarted.Add(ctx, 1)
}

func (m *MetricsCollector) RecordTransactionSuccess(duration float64) {
	ctx := context.Background()
	m.transactionsSuccess.Add(ctx, 1)
	m.transactionDuration.Record(ctx, duration)
}

func (m *MetricsCollector) RecordTransactionFailed(duration float64) {
	ctx := context.Background()
	m.transactionsFailed.Add(ctx, 1)
	m.transactionDuration.Record(ctx, duration)
}

func (m *MetricsCollector) RecordDBOperation(success bool, duration float64) {
	ctx := context.Background()
	m.dbOperationsTotal.Add(ctx, 1)
	m.dbOperationDuration.Record(ctx, duration)
	if !success {
		m.dbOperationErrors.Add(ctx, 1)
	}
}
