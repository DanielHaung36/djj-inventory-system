package service

import (
	"djj-inventory-system/internal/model"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// RecordAudit 在 audit_histories 里写一条记录
func RecordAudit(db *gorm.DB, table model.AuditedTableEnum, recordID int, changedBy int, operation string, payload interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ah := model.AuditedHistory{
		TableName: table,
		RecordID:  recordID,
		ChangedBy: changedBy,
		Operation: operation,
		Payload:   raw,
		ChangedAt: time.Now(),
	}
	return db.Create(&ah).Error
}
