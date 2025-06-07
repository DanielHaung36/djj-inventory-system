package importdata

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"djj-inventory-system/internal/models"
)

func getCategoryAndTypeCode(category, subCategory, thirdLevel, model string) (string, int) {
	// 所有字段统一大写处理
	cat := strings.ToUpper(category)
	sub := strings.ToUpper(subCategory)
	thd := strings.ToUpper(thirdLevel)
	mdl := strings.ToUpper(model)

	switch {
	// 主机：装载机
	case strings.Contains(cat, "主机") && strings.Contains(sub, "装载机"):
		return "MH", 1
	// 主机：叉车
	case strings.Contains(cat, "主机") && strings.Contains(sub, "叉车"):
		return "MH", 2
	// 主机：挖掘机
	case strings.Contains(cat, "主机") && strings.Contains(sub, "挖掘机"):
		return "MH", 3
	// 主机：滑移机
	case strings.Contains(cat, "主机") && strings.Contains(sub, "滑移机"):
		return "MH", 4
	// 主机：剪叉车
	case strings.Contains(cat, "主机") && strings.Contains(sub, "剪叉"):
		return "MH", 5

	// 配件：装载机配件
	case strings.Contains(cat, "配件") && strings.Contains(sub, "装载机"):
		if strings.Contains(thd, "滤芯") {
			return "PT", 1
		}
		return "PT", 10 // 装载机其他配件

	// 配件：挖掘机驾驶室
	case strings.Contains(cat, "配件") && strings.Contains(sub, "挖掘机") && strings.Contains(thd, "驾驶室"):
		return "PT", 2

	// 配件：叉车配件
	case strings.Contains(cat, "配件") && strings.Contains(sub, "叉车"):
		return "PT", 3

	// 属具：装载机铲斗
	case strings.Contains(cat, "属具") && strings.Contains(sub, "装载机") && strings.Contains(thd, "铲斗"):
		return "AT", 1
	// 属具：装载机抓斗
	case strings.Contains(cat, "属具") && strings.Contains(sub, "装载机") && strings.Contains(thd, "抓斗"):
		return "AT", 2
	// 属具：挖掘机
	case strings.Contains(cat, "属具") && strings.Contains(sub, "挖掘机"):
		return "AT", 3
	// 属具：滑移机
	case strings.Contains(cat, "属具") && strings.Contains(sub, "滑移机"):
		return "AT", 4
	// 属具：叉车
	case strings.Contains(cat, "属具") && strings.Contains(sub, "叉车"):
		return "AT", 5

	// 工具
	case strings.Contains(cat, "工具") && strings.Contains(sub, "千斤顶"):
		return "TL", 1
	case strings.Contains(cat, "工具") && strings.Contains(sub, "扳手"):
		return "TL", 2

	// IT
	case strings.Contains(cat, "IT") && strings.Contains(sub, "笔记本"):
		return "IT", 1
	case strings.Contains(cat, "IT") && (strings.Contains(sub, "服务器") || strings.Contains(sub, "打印机")):
		return "IT", 2

	// 其他
	default:
		return "OT", 1 // 归为其他
	}
}

// parseFloat 将字符串转换为 float64，出错时返回 0
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ImportProducts 从 Excel 导入 products
// 假设列顺序：
// 0:djj_code, 1:name_cn, 2:name_en, 3:manufacturer,
// 4:manufacturer_code, 5:supplier, 6:model,
// 7:category_name, 8:price, 9:rrp_price, 10:currency
//DJJ code	状态	供货商	厂家唯一代码	产品类别	产品子类别	三级类别	品名中文	品名英文	适配机型/规格	标准Standards	单位	货币	销售RRP价格	Standard Warranty	备注	SourceID	DJJ唯一识别代码	图片 Photo	尺寸 Dimension (L*W*H - mm)	重量 Weight (kg)	起重 Lift Capacity (KG)	起升高度 Lift Height (mm)	动力源 Power Source	其他配置 Other Specs	质保 Warranty	该Code 专属营销销售信息汇总	主机类别通用营销销售信息汇总	产品知识拓展资料（厂家培训，故障排除）	SYD Stock	PER Stock	BNE Stock	📋 待发运物品清单 updated on Oct. 22	最后更新时间	最近修改人	该code专属网页链接

func ImportProducts(db *gorm.DB, f *excelize.File) error {
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
		if len(row) < 11 {
			// 列数不足，跳过或根据需求报错
			continue
		}

		// 查找分类
		var cat models.ProductCategory
		if err := tx.Where("name = ?", row[4]).First(&cat).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("第%d行: 找不到分类 '%s': %w", i+1, row[7], err)
		}
		// 构造产品结构体
		prod := models.Product{
			DjjCode:      row[0],
			NameCn:       row[5],
			NameEn:       sql.NullString{String: row[6], Valid: true},
			Manufacturer: sql.NullString{String: row[1], Valid: true},
			Supplier:     sql.NullString{String: row[1], Valid: true},
			Model:        sql.NullString{String: row[7], Valid: true},
			CategoryID:   sql.NullInt64{Int64: int64(cat.ID), Valid: true},
			Price:        parseFloat(row[8]),
			RrpPrice:     sql.NullFloat64{Float64: 0, Valid: true},
			Currency:     models.CurrencyCodeEnumAud,
			Status:       models.ProductStatusEnumDraft,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 插入或更新（按 djj_code 冲突时更新所有字段）
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "djj_code"}},
			UpdateAll: true,
		}).Create(&prod).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("第%d行插入产品失败: %w", i+1, err)
		}
	}

	return tx.Commit().Error
}
