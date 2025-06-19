package audit

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"djj-inventory-system/internal/model"
)

// Recorder 是审计接口
type Recorder interface {
	Record(ctx context.Context, refType model.AuditedTableEnum, refID uint, op string, payload interface{}) error
}

// GormAuditor 把日志写到 audited_history
type GormAuditor struct {
	db *gorm.DB
}

func NewGormAuditor(db *gorm.DB) Recorder {
	return &GormAuditor{db: db}
}

func (a *GormAuditor) Record(ctx context.Context, refType model.AuditedTableEnum, refID uint, op string, payload interface{}) error {
	var raw json.RawMessage

	// 如果 payload 本来就是 RawMessage，就直接用
	switch v := payload.(type) {
	case json.RawMessage:
		raw = v
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		raw = b
	}

	// 从 context 拿当前用户
	uidVal := ctx.Value(model.ContextUserIDKey)
	userID, ok := uidVal.(uint)
	if !ok {
		return errors.New("audit: missing user ID in context")
	}

	hist := model.AuditedHistory{
		TableName: refType,
		RecordID:  int(refID),
		StoreID:   0, // 如果有门店，可以改成从 ctx 或 payload 里解析
		ChangedBy: int(userID),
		Operation: op,
		Payload:   raw,
	}

	return a.db.WithContext(ctx).Create(&hist).Error
}
