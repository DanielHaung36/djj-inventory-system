package audit

import (
	"context"
	"djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/common"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// Recorder 是审计接口
type Recorder interface {
	Record(ctx context.Context, refType audit.AuditedTableEnum, refID uint, op string, payload interface{}) error
}

// GormAuditor 把日志写到 audited_history
type GormAuditor struct {
	db *gorm.DB
}

func NewGormAuditor(db *gorm.DB) Recorder {
	return &GormAuditor{db: db}
}

func (a *GormAuditor) Record(ctx context.Context, refType audit.AuditedTableEnum, refID uint, op string, payload interface{}) error {
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
	uidVal := ctx.Value(common.ContextUserIDKey)
	userID, ok := uidVal.(uint)
	if !ok {
		return errors.New("audit: missing user ID in context")
	}

	hist := audit.AuditedHistory{
		TableName: refType,
		RecordID:  int(refID),
		StoreID:   0, // 如果有门店，可以改成从 ctx 或 payload 里解析
		ChangedBy: int(userID),
		Operation: op,
		Payload:   raw,
	}

	return a.db.WithContext(ctx).Create(&hist).Error
}

// MockRecorder 是一个用于测试的 mock audit recorder
type MockRecorder struct{}

func (m *MockRecorder) Record(ctx context.Context, userID uint, table common.AuditedTable, recordID uint, action common.AuditAction, old, new any) error {
	// 在测试中，这个方法是空的，因为它不需要真的记录
	return nil
}
