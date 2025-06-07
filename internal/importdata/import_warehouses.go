package importdata

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"djj-inventory-system/internal/models"
)

var RawJSON = []byte(`
{
  "regions": [
    { "name": "NSW" },
    { "name": "QLD" },
    { "name": "WA" }
  ],
  "stores": [
    { "code": "SYD", "name": "Sydney",   "region": "NSW" },
    { "code": "BNE", "name": "Brisbane", "region": "QLD" },
    { "code": "PER", "name": "Perth",    "region": "WA" }
  ],
  "warehouses": [
    { "name": "SYD", "location": "Sydney, NSW",   "region": "NSW" },
    { "name": "BNE", "location": "Brisbane, QLD", "region": "QLD" },
    { "name": "PER", "location": "Perth, WA",     "region": "WA" }
  ]
}
`)

type locationJSON struct {
	Regions []struct {
		Name string `json:"name"`
	} `json:"regions"`
	Stores []struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Region string `json:"region"`
	} `json:"stores"`
	Warehouses []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
		Region   string `json:"region"`
	} `json:"warehouses"`
}

// ImportFromJSON 从 JSON 创建 regions／stores／warehouses 及关联
func ImportFromJSON(db *gorm.DB, raw []byte) error {

	var data locationJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 缓存名称到 ID
	regionCache := map[string]uint{}
	// 先建区
	for _, r := range data.Regions {
		reg := models.Region{Name: r.Name}
		tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&reg)
		// 如果第一次插入，reg.ID 自动填；否则再查一次
		if reg.ID == 0 {
			tx.Where("name = ?", r.Name).First(&reg)
		}
		regionCache[r.Name] = uint(reg.ID)
	}

	// 建门店
	for _, s := range data.Stores {
		rid, ok := regionCache[s.Region]
		if !ok {
			tx.Rollback()
			return fmt.Errorf("未知的 region %q for store %q", s.Region, s.Code)
		}
		st := models.Store{
			Code:     s.Code,
			Name:     s.Name,
			RegionID: int(uint(rid)),
		}
		tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			UpdateAll: true,
		}).Create(&st)
	}

	// 建仓库并关联
	for _, w := range data.Warehouses {
		rid, ok := regionCache[w.Region]
		if !ok {
			tx.Rollback()
			return fmt.Errorf("未知的 region %q for warehouse %q", w.Region, w.Name)
		}
		wh := models.Warehouse{
			Name:     w.Name,
			Location: sql.NullString{String: w.Location, Valid: true},
		}
		tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&wh)
		if wh.ID == 0 {
			tx.Where("name = ?", w.Name).First(&wh)
		}
		// region_warehouses 关联
		err := tx.Exec(
			"INSERT INTO region_warehouses(region_id, warehouse_id) VALUES(?, ?) ON CONFLICT DO NOTHING",
			rid, wh.ID,
		).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("关联 region %d 和 warehouse %d 失败: %w", rid, wh.ID, err)
		}
	}

	return tx.Commit().Error
}
