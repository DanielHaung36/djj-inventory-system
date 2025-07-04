package catalog

import (
	"djj-inventory-system/internal/model/company"
	"djj-inventory-system/internal/model/rbac"
	"time"

	"gorm.io/gorm"
)

// Store 对应数据库表 stores
type Store struct {
	ID        uint            `gorm:"primaryKey" json:"id"`                          // 门店ID
	Code      string          `gorm:"size:50;unique;not null" json:"code"`           // 门店编码
	Name      string          `gorm:"size:100;not null" json:"name"`                 // 门店名称
	RegionID  uint            `gorm:"not null" json:"regionId"`                      // 所属地区ID
	Region    Region          `gorm:"foreignKey:RegionID" json:"region,omitempty"`   // 关联地区
	CompanyID uint            `gorm:"not null" json:"companyId"`                     // 所属公司ID
	Company   company.Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"` // 关联公司
	Address   string          `gorm:"size:255" json:"address"`                       // 地址
	ManagerID uint            `json:"managerId,omitempty"`                           // 负责人ID
	Manager   rbac.User       `gorm:"foreignKey:ManagerID" json:"manager,omitempty"` // 负责人
	Version   int64           `gorm:"not null;default:1" json:"version"`             // 乐观锁
	CreatedAt time.Time       `json:"createdAt"`                                     // 创建时间
	UpdatedAt time.Time       `json:"updatedAt"`                                     // 更新时间
	IsDeleted bool            `gorm:"not null;default:false" json:"isDeleted"`       // 软删除标记
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`                                // GORM 软删除字段（可选）
}

// TableName 指定这张表在数据库里的名字
func (Store) TableName() string {
	return "stores"
}
