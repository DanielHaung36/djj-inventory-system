package catalog // 或者 internal/model

import (
	"time"
)

// Attachment 对应数据库表 attachments，支持多种 ref_type/ref_id 的多态关联
type Attachment struct {
	ID         uint      `gorm:"primaryKey"              json:"id"`
	FileName   string    `gorm:"size:255;not null"       json:"fileName"`
	FileType   string    `gorm:"size:100;not null"       json:"fileType"`
	FileSize   int       `json:"fileSize"`
	URL        string    `gorm:"type:text;not null"      json:"url"`
	UploadedBy uint      `json:"uploadedBy"` // gorm 默认会映射到 uploaded_by
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploadedAt"`
	RefType    string    `gorm:"size:20;not null"        json:"refType"` // 比如 "product", "quote", "order"...
	RefID      uint      `gorm:"not null"               json:"refId"`
}

// TableName 明确指定表名
func (Attachment) TableName() string {
	return "attachments"
}
