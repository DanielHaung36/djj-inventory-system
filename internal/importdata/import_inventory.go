package importdata

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"djj-inventory-system/internal/models"
)

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
func parseDate(s string) time.Time {
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Now()
}

// ImportInventoryFromExcel 按你那张表的列来导入并更新库存
func ImportInventoryFromExcel(db *gorm.DB, f *excelize.File) error {
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("读取 Excel 行失败: %w", err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}
		if len(row) < 15 {
			continue // 列数不足
		}

		// 1) 找到产品
		var prod models.Product
		if err := tx.Where("djj_code = ?", row[0]).First(&prod).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("第 %d 行: 未找到产品 %s: %w", i+1, row[0], err)
		}

		// 2) 找到仓库（假设仓库 code 就是 row[1]，如果是 name 则改 Where）
		var wh models.Warehouse
		if err := tx.Where("name = ?", row[1]).First(&wh).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("第 %d 行: 未找到仓库 %s: %w", i+1, row[1], err)
		}

		// 3) 读取数量与日期
		onHand := parseInt(row[8])
		reserved := parseInt(row[9])
		actualCount := parseInt(row[11])
		// 优先使用盘库日期更新，如果没值则用最后修改时间
		updatedAt := parseDate(row[12])
		if t2 := parseDate(row[16]); !t2.IsZero() {
			updatedAt = t2
		}

		// 4) 获取或创建 inventory
		var inv models.Inventory
		err = tx.Where("product_id = ? AND warehouse_id = ?", prod.ID, wh.ID).
			First(&inv).Error
		if err == gorm.ErrRecordNotFound {
			inv = models.Inventory{
				ProductID:        prod.ID,
				WarehouseID:      wh.ID,
				OnHand:           onHand,
				ReservedForOrder: reserved,
				UpdatedAt:        updatedAt,
			}
			if err := tx.Create(&inv).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("第 %d 行: 创建库存行失败: %w", i+1, err)
			}
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("第 %d 行: 查询库存失败: %w", i+1, err)
		} else {
			// 5) 更新库存并计算差值
			delta := actualCount - inv.OnHand
			inv.OnHand = onHand
			inv.ReservedForOrder = reserved
			inv.UpdatedAt = updatedAt

			if err := tx.Save(&inv).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("第 %d 行: 更新库存失败: %w", i+1, err)
			}

			// 6) 写盘库日志
			log := models.InventoryLog{
				InventoryID: inv.ID,
				ChangeType:  "stocktake", // 盘库
				Quantity:    delta,
				Operator:    row[15],                                      // 最后修改人
				Remark:      sql.NullString{String: row[13], Valid: true}, // 盘库备注
				CreatedAt:   updatedAt,
			}
			if err := tx.Create(&log).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("第 %d 行: 写日志失败: %w", i+1, err)
			}
		}
	}

	return tx.Commit().Error
}
