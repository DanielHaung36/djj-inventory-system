package importdata

import (
	"database/sql"
	"djj-inventory-system/internal/models"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ImportCategories 从 Excel 导入 product_categories（支持三级分类）
func ImportCategories(db *gorm.DB, f *excelize.File) error {
	// 获取表名并读取所有行
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("读取 Excel 行失败: %w", err)
	}

	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 本地缓存，避免重复查询/创建
	lvl1Cache := make(map[string]uint)
	lvl2Cache := make(map[string]uint)
	lvl3Cache := make(map[string]uint)

	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}
		if len(row) < 5 {
			continue // 列数不足，跳过
		}

		// 按列索引提取：row[2]=一级，row[3]=二级，row[4]=三级
		names := []string{row[2], row[3], row[4]}
		var parentID uint

		// 依次处理三级分类
		for level, name := range names {
			if name == "" {
				break // 若某层为空则后续不处理
			}

			// 根据层级选择缓存和 key
			var cache map[string]uint
			var key string
			switch level {
			case 0:
				cache = lvl1Cache
				key = name
			case 1:
				cache = lvl2Cache
				key = fmt.Sprintf("%d|%s", parentID, name)
			case 2:
				cache = lvl3Cache
				key = fmt.Sprintf("%d|%s", parentID, name)
			}

			var catID uint
			if id, ok := cache[key]; ok {
				// 已缓存，直接复用
				catID = id
			} else {
				// 创建或获取已存在分类
				cat := models.ProductCategory{
					Name:     name,
					ParentID: sql.NullInt64{Int64: int64(parentID), Valid: level > 0},
				}
				tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "name"}, {Name: "parent_id"}},
					DoNothing: true,
				}).Create(&cat)

				if cat.ID == 0 {
					// 已存在，查询其 ID
					tx.Where("name = ? AND parent_id = ?", name, cat.ParentID).
						First(&cat)
				}
				catID = uint(cat.ID)
				cache[key] = catID
			}

			// 本层创建后作为下一层的父 ID
			parentID = catID
		}
	}

	return tx.Commit().Error
}
