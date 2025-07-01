package database

import (
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/rbac"
	"djj-inventory-system/internal/model/sales"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func SeedTestData(db *gorm.DB) error {
	rand.Seed(time.Now().UnixNano())

	// 1. Seed Customers
	var customers []catalog.Customer
	for i := 1; i <= 5; i++ {
		c := catalog.Customer{
			StoreID: uint(rand.Intn(3) + 1),
			Name:    fmt.Sprintf("Customer %d", i),
			Phone:   fmt.Sprintf("0400%05d", i),
			Email:   fmt.Sprintf("cust%d@example.com", i),
			Address: fmt.Sprintf("%d Test Street, WA", i),
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
		customers = append(customers, c)
	}

	// 2. Seed Products
	var products []catalog.Product
	for i := 1; i <= 20; i++ {
		p := catalog.Product{
			DJJCode:           fmt.Sprintf("P%05d", i),
			NameCN:            fmt.Sprintf("产品%d", i),
			NameEN:            fmt.Sprintf("Product %d", i),
			Manufacturer:      fmt.Sprintf("厂商%d", rand.Intn(10)),
			ManufacturerCode:  fmt.Sprintf("MFC%03d", rand.Intn(100)),
			Supplier:          fmt.Sprintf("供应商%d", rand.Intn(10)),
			Model:             fmt.Sprintf("MDL%03d", rand.Intn(100)),
			Price:             rand.Float64() * 10000,
			RRPPrice:          rand.Float64() * 12000,
			Currency:          "AUD",
			Status:            "draft",
			ApplicationStatus: "open",
			ProductType:       "others",
			Version:           1,
		}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
		products = append(products, p)
	}

	// 3. Seed Quotes + QuoteItems
	var quotes []sales.Quote
	for i := 1; i <= 5; i++ {
		q := sales.Quote{
			StoreID:      1,
			CustomerID:   customers[rand.Intn(len(customers))].ID,
			QuoteNumber:  fmt.Sprintf("QTE-%04d", i),
			SalesRepUser: rbac.User{Username: "Mark Wang"},
			QuoteDate:    time.Now(),
			Currency:     "AUD",
			Status:       "pending",
		}
		if err := db.Create(&q).Error; err != nil {
			return err
		}

		var subTotal float64
		// 每张报价 1–4 条明细
		cnt := rand.Intn(4) + 1
		for j := 0; j < cnt; j++ {
			prod := products[rand.Intn(len(products))]
			qty := rand.Intn(5) + 1
			unitPrice := prod.Price
			totalPrice := unitPrice * float64(qty)

			qi := sales.QuoteItem{
				QuoteID:     q.ID,
				ProductID:   &prod.ID,
				Description: prod.NameEN,
				Quantity:    qty,
				Unit:        "ea",
				UnitPrice:   unitPrice,
				Discount:    0,
				TotalPrice:  totalPrice,
				GoodsNature: "contract",
			}
			if err := db.Create(&qi).Error; err != nil {
				return err
			}
			subTotal += totalPrice
		}

		gst := subTotal * 0.1
		q.SubTotal = subTotal
		q.GSTTotal = gst
		q.TotalAmount = subTotal + gst
		if err := db.Save(&q).Error; err != nil {
			return err
		}
		quotes = append(quotes, q)
	}

	// 4. Seed Orders + OrderItems
	var orders []sales.Order
	for i := 1; i <= 5; i++ {
		src := quotes[rand.Intn(len(quotes))]
		o := sales.Order{
			QuoteID:         &src.ID,
			StoreID:         src.StoreID,
			CustomerID:      src.CustomerID,
			OrderNumber:     fmt.Sprintf("ORD-%04d", i),
			OrderDate:       time.Now(),
			Currency:        "AUD",
			ShippingAddress: customers[0].Address,
			Status:          "draft",
		}
		if err := db.Create(&o).Error; err != nil {
			return err
		}

		// 把 quote_items 拷贝成 order_items
		var qis []sales.QuoteItem
		if err := db.Where("quote_id = ?", src.ID).Find(&qis).Error; err != nil {
			return err
		}
		for _, qi := range qis {
			oi := sales.OrderItem{
				OrderID:   o.ID,
				ProductID: *qi.ProductID,
				Quantity:  qi.Quantity,
				UnitPrice: qi.UnitPrice,
			}
			if err := db.Create(&oi).Error; err != nil {
				return err
			}
		}
		orders = append(orders, o)
	}

	// 5. Seed PickingLists + PickingListItems
	for i, o := range orders {
		pl := sales.PickingList{
			OrderID:         o.ID,
			PickingNumber:   fmt.Sprintf("PKG-%04d", i+1),
			DeliveryAddress: o.ShippingAddress,
			Status:          "draft",
		}
		if err := db.Create(&pl).Error; err != nil {
			return err
		}

		var ois []sales.OrderItem
		if err := db.Where("order_id = ?", o.ID).Find(&ois).Error; err != nil {
			return err
		}
		for _, oi := range ois {
			pli := sales.PickingListItem{
				PickingListID: pl.ID,
				ProductID:     oi.ProductID,
				Quantity:      oi.Quantity,
				Location:      "Main Warehouse",
			}
			if err := db.Create(&pli).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
