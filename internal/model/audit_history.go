package model

import (
	"encoding/json"
	"time"
)

// 对应 DB 中 audited_history 表
type AuditedHistory struct {
	HistoryID int              `gorm:"column:history_id;primaryKey"`
	TableName AuditedTableEnum `gorm:"column:table_name"`
	RecordID  int              `gorm:"column:record_id"`
	StoreID   int              `gorm:"column:store_id"`
	ChangedBy int              `gorm:"column:changed_by"`
	Operation string           `gorm:"column:operation"`
	Payload   json.RawMessage  `gorm:"column:payload"`
	ChangedAt time.Time        `gorm:"column:changed_at"` // 或 time.Time
}

// 枚举：把所有需要审计的表名都加进来
type AuditedTableEnum string

const (
	AuditedTableUsers           AuditedTableEnum = "users"
	AuditedTableRoles           AuditedTableEnum = "roles"
	AuditedTablePermissions     AuditedTableEnum = "permissions"
	AuditedTableUserRoles       AuditedTableEnum = "user_roles"
	AuditedTableRolePermissions AuditedTableEnum = "role_permissions"
	AuditedTableProducts        AuditedTableEnum = "products"
	AuditedTableQuotes          AuditedTableEnum = "quotes"
	AuditedTableQuoteItems      AuditedTableEnum = "quote_items"
	AuditedTableOrders          AuditedTableEnum = "orders"
	AuditedTableOrderItems      AuditedTableEnum = "order_items"
	AuditedTableInventory       AuditedTableEnum = "inventory"
	AuditedTableInventoryLogs   AuditedTableEnum = "inventory_logs"
	// ……按需继续添加
)
