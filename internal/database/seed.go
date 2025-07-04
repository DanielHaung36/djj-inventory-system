package database

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/company"
	"djj-inventory-system/internal/model/rbac"
	"djj-inventory-system/internal/model/sales"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedTestData 初始化测试数据：公司 → 区域 → 门店 → 客户 → 产品 → 报价 → 订单 → 拣货单
type Seeder struct {
	db *gorm.DB
}

// NewSeeder 返回新的 Seeder
func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

// Run 执行所有 Seed
func (s *Seeder) Run() error {
	rand.Seed(time.Now().UnixNano())

	// 0. Seed Companies
	companies := []company.Company{
		{Code: "DJJ_PERTH", Name: "DJJ PERTH PTY LTD", Email: "sales@djjequipment.com.au", Phone: "1800 355 388", Website: "https://djjequipment.com.au", ABN: "95 663 874 664", Address: "56 Clavering Road Bayswater, WA Australia 6053", IsDefault: true, BSB: "956638"},
		{Code: "DJJ_BRISBANE", Name: "DJJ BRISBANE PTY LTD", Email: "sales.brisbane@djjequipment.com.au", Phone: "1800 355 389", Website: "https://brisbane.djjequipment.com.au", ABN: "12 345 678 901", Address: "123 Queen Street, Brisbane, QLD 4000", IsDefault: false, BSB: "123678"},
		{Code: "DJJ_SYDNEY", Name: "DJJ SYDNEY PTY LTD", Email: "sales.sydney@djjequipment.com.au", Phone: "1800 355 390", Website: "https://sydney.djjequipment.com.au", ABN: "98 765 432 109", Address: "456 George Street, Sydney, NSW 2000", IsDefault: false, BSB: "987654"},
	}
	var seededCompanies []company.Company
	for i := range companies {
		c := companies[i]
		now := time.Now()
		c.CreatedAt = now
		c.UpdatedAt = now
		if err := s.db.Where(company.Company{Code: c.Code}).FirstOrCreate(&c).Error; err != nil {
			return fmt.Errorf("seed company %s: %w", c.Code, err)
		}
		seededCompanies = append(seededCompanies, c)
	}

	// 1. Seed Regions
	var regions []catalog.Region
	regionNames := []string{"Perth", "Brisbane", "Sydney"}
	for idx, name := range regionNames {
		r := catalog.Region{Name: name, CompanyID: seededCompanies[idx].ID}
		if err := s.db.Where("name = ?", r.Name).FirstOrCreate(&r).Error; err != nil {
			return fmt.Errorf("seed region %s: %w", r.Name, err)
		}
		regions = append(regions, r)
	}

	// 1.5. Seed Warehouses （每个 Region 多个仓库）
	var warehouses []catalog.Warehouse
	for _, r := range regions {
		// 你可以根据 Region 名称或其它策略，生成多个仓库配置
		warehouseNames := []string{
			fmt.Sprintf("%s Main Warehouse", r.Name),
			fmt.Sprintf("%s Overflow Warehouse", r.Name),
		}
		for _, wn := range warehouseNames {
			w := catalog.Warehouse{
				Name:      wn,
				Location:  fmt.Sprintf("%s Location", wn),
				Version:   1,
				IsDeleted: false,
			}
			if err := s.db.
				Where("name = ?", w.Name).
				FirstOrCreate(&w).Error; err != nil {
				return fmt.Errorf("seed warehouse %s: %w", w.Name, err)
			}
			warehouses = append(warehouses, w)

			// 同时插入 Region ↔ Warehouse 的关联
			rw := catalog.RegionWarehouse{
				RegionID:    r.ID,
				WarehouseID: w.ID,
			}
			if err := s.db.
				Where("region_id = ? AND warehouse_id = ?", rw.RegionID, rw.WarehouseID).
				FirstOrCreate(&rw).Error; err != nil {
				return fmt.Errorf(
					"seed region_warehouse for region %d warehouse %d: %w",
					r.ID, w.ID, err,
				)
			}
		}
	}

	// 2. Seed Stores
	var stores []catalog.Store
	// 1) 先只创建好所有的 Store，不带 manager_id
	for _, r := range regions {
		code := strings.ToUpper(r.Name[:3]) + "_STORE"
		st := catalog.Store{
			Code:      code,
			Name:      r.Name + " Store",
			RegionID:  r.ID,
			CompanyID: r.CompanyID,
			// **不要** 写 ManagerID
		}
		// Omit 掉 manager_id，保证第一次插入时不会带 manager_id=0
		if err := s.db.
			Omit("manager_id").
			Where("code = ? AND company_id = ?", st.Code, st.CompanyID).
			FirstOrCreate(&st).Error; err != nil {
			return fmt.Errorf("seed store %s: %w", st.Code, err)
		}
		stores = append(stores, st)
	}

	// 2) 再给每个 Store 创建一个 sales_leader 用户
	for _, st := range stores {
		username := "sales_leader_" + strings.ToLower(st.Code)
		hash, _ := bcrypt.GenerateFromPassword([]byte("qq123456"), bcrypt.DefaultCost)
		usr := rbac.User{
			Username:     username,
			Email:        username + "@example.com",
			PasswordHash: string(hash),
			StoreID:      st.ID,
		}
		if err := s.db.
			Where("username = ?", usr.Username).
			FirstOrCreate(&usr).
			Error; err != nil {
			return fmt.Errorf("seed user %s: %w", usr.Username, err)
		}

		// 3) 拿到 usr.ID 之后，单独更新这条 store 的 manager_id
		if err := s.db.
			Model(&catalog.Store{}).
			Where("id = ?", st.ID).
			Update("manager_id", usr.ID).
			Error; err != nil {
			return fmt.Errorf("assign manager to store %d: %w", st.ID, err)
		}
	}

	// 3. Seed Customers
	for idx, st := range stores {
		for j := 1; j <= 2; j++ {
			n := idx*2 + j
			customer := catalog.Customer{
				StoreID: st.ID,
				Type:    "retail",
				Company: "Wyndham Youth Aboriginal\nCorporation",
				Name:    fmt.Sprintf("Customer %d", n),
				Phone:   fmt.Sprintf("0400%05d", n),
				Contact: fmt.Sprintf("%s Wyndham Youth Aboriginal", st.Name),
				Email:   fmt.Sprintf("cust%d@%s.com", n, strings.ToLower(st.Code)),
				Address: fmt.Sprintf("%d Test Street, %s", n, st.Name),
				ABN:     fmt.Sprintf("1100%05d", n),
				Version: 1,
			}
			if err := s.db.
				Where("store_id = ? AND name = ?", customer.StoreID, customer.Name).
				FirstOrCreate(&customer, customer).Error; err != nil {
				return fmt.Errorf("seed customer %s: %w", customer.Name, err)
			}
		}
	}
	for idx, st := range stores {
		for j := 1; j <= 2; j++ {
			n := idx*2 + j
			customer := catalog.Customer{
				StoreID: st.ID,
				Type:    "retail",
				Contact: fmt.Sprintf("%s Wyndham Youth Aboriginal", st.Name),
				Company: "Dnaile Youth Aboriginal\nCorporation",
				Name:    fmt.Sprintf("Customer %d", n),
				Phone:   fmt.Sprintf("0400%05d", n),
				Email:   fmt.Sprintf("cust%d@%s.com", n, strings.ToLower(st.Code)),
				Address: fmt.Sprintf("%d Test Street, %s", n, st.Name),
				ABN:     fmt.Sprintf("1100%05d", n),
				Version: 1,
			}
			if err := s.db.
				Where("store_id = ? AND name = ?", customer.StoreID, customer.Name).
				FirstOrCreate(&customer, customer).Error; err != nil {
				return fmt.Errorf("seed customer %s: %w", customer.Name, err)
			}
		}
	}
	for idx, st := range stores {
		for j := 1; j <= 2; j++ {
			n := idx*2 + j
			customer := catalog.Customer{
				StoreID: st.ID,
				Type:    "retail",

				Name:    fmt.Sprintf("Customer %d", n),
				Phone:   fmt.Sprintf("0400%05d", n),
				Email:   fmt.Sprintf("cust%d@%s.com", n, strings.ToLower(st.Code)),
				Address: fmt.Sprintf("%d Test Street, %s", n, st.Name),
				ABN:     fmt.Sprintf("1100%05d", n),
				Version: 1,
			}
			// 根据 store_id 和 name 判重
			if err := s.db.Where("store_id = ? AND name = ?", customer.StoreID, customer.Name).
				FirstOrCreate(&customer, customer).Error; err != nil {
				return fmt.Errorf("seed customer %s: %w", customer.Name, err)
			}
		}
	}
	for idx, st := range stores {
		for j := 1; j <= 2; j++ {
			n := idx*2 + j
			c := catalog.Customer{StoreID: st.ID, Type: "retail", Name: fmt.Sprintf("Customer %d", n), Phone: fmt.Sprintf("0400%05d", n), Email: fmt.Sprintf("cust%d@%s.com", n, strings.ToLower(st.Code)), Address: fmt.Sprintf("%d Test Street, %s", n, st.Name), ABN: fmt.Sprintf("1100%05d", n), Version: 1}
			if err := s.db.Where("store_id = ? AND name = ?", c.StoreID, c.Name).FirstOrCreate(&c).Error; err != nil {
				return fmt.Errorf("seed customer %s: %w", c.Name, err)
			}
		}
	}

	// 4. Seed Products
	var products []catalog.Product
	for i := 1; i <= 20; i++ {
		p := catalog.Product{DJJCode: fmt.Sprintf("P%05d", i), NameCN: fmt.Sprintf("产品%d", i), NameEN: fmt.Sprintf("Product %d", i), Manufacturer: fmt.Sprintf("厂商%d", rand.Intn(10)), ManufacturerCode: fmt.Sprintf("MFC%03d", rand.Intn(100)), Supplier: fmt.Sprintf("供应商%d", rand.Intn(10)), Model: fmt.Sprintf("MDL%03d", rand.Intn(100)), Price: rand.Float64() * 10000, RRPPrice: rand.Float64() * 12000, Currency: "AUD", Category: catalog.CategoryMachine, Status: "draft", ApplicationStatus: "open", ProductType: "others", Version: 1}
		if err := s.db.Where("djj_code = ?", p.DJJCode).FirstOrCreate(&p).Error; err != nil {
			return fmt.Errorf("seed product %s: %w", p.DJJCode, err)
		}
		products = append(products, p)
	}

	// 5. Seed Quotes
	var quotes []sales.Quote
	for i := 1; i <= 5; i++ {
		// 随机选客户
		var cust catalog.Customer
		if err := s.db.Order("RANDOM()").First(&cust).Error; err != nil {
			return err
		}
		// 加载对应门店
		var st catalog.Store
		if err := s.db.First(&st, cust.StoreID).Error; err != nil {
			return fmt.Errorf("load store %d: %w", cust.StoreID, err)
		}
		// 使用动态用户名查找对应门店的 sales_leader
		leaderUsername := fmt.Sprintf("sales_leader_%s", strings.ToLower(st.Code))
		var rep rbac.User
		if err := s.db.Where("username = ? AND store_id = ?", leaderUsername, st.ID).
			First(&rep).Error; err != nil {
			return fmt.Errorf("load sales rep %s for store %d: %w", leaderUsername, st.ID, err)
		}

		// 创建报价
		q := sales.Quote{
			StoreID:     st.ID,
			CompanyID:   st.CompanyID,
			CustomerID:  cust.ID,
			QuoteNumber: fmt.Sprintf("QTE-%04d", i),
			SalesRepID:  rep.ID,
			QuoteDate:   time.Now(),
			Currency:    "AUD",
			Status:      "pending",
		}
		if err := s.db.Create(&q).Error; err != nil {
			return fmt.Errorf("create quote %s: %w", q.QuoteNumber, err)
		}

		// 生成明细并统计小计
		var subTotal float64
		cnt := rand.Intn(4) + 1
		for j := 0; j < cnt; j++ {
			var prod catalog.Product
			if err := s.db.Order("RANDOM()").First(&prod).Error; err != nil {
				return fmt.Errorf("pick random product: %w", err)
			}
			qty := rand.Intn(5) + 1
			total := prod.Price * float64(qty)
			qi := sales.QuoteItem{QuoteID: q.ID, ProductID: &prod.ID, Description: prod.NameEN, Quantity: qty, Unit: "ea", UnitPrice: prod.Price, Discount: 0, TotalPrice: total, GoodsNature: "contract"}
			if err := s.db.Create(&qi).Error; err != nil {
				return fmt.Errorf("create quote item: %w", err)
			}
			subTotal += total
		}

		// 更新金额字段
		q.SubTotal = subTotal
		q.GSTTotal = subTotal * 0.10
		q.TotalAmount = q.SubTotal + q.GSTTotal
		if err := s.db.Save(&q).Error; err != nil {
			return fmt.Errorf("update quote totals: %w", err)
		}
		quotes = append(quotes, q)
	}

	return nil
}
