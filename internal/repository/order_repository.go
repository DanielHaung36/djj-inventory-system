package repository

import (
	"context"
	"djj-inventory-system/internal/model/sales"
	"errors"

	"gorm.io/gorm"
)

// OrderRepository 封装对 orders 表的访问
type OrderRepository struct {
	DB *gorm.DB
}

// NewOrderRepository 构造函数
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

// FindByID 根据主键读取订单（并且把 Store、Customer、Items，以及每个 Item 的 Product 一起 Preload 进来）
func (r *OrderRepository) FindByID(ctx context.Context, id uint) (*sales.Order, error) {
	var o sales.Order
	err := r.DB.WithContext(ctx).
		Preload("Store").
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("SalesRepUser").
		Preload("Items.Product").
		First(&o, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

// Create 在数据库中新增一条订单记录（包含关联的 Items）
func (r *OrderRepository) Create(ctx context.Context, order *sales.Order) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		// 假设 front-end 已经把 Items 放在 order.Items 里
		return nil
	})
}

// Update 更新一条订单（及其明细）
func (r *OrderRepository) Update(ctx context.Context, order *sales.Order) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// 如果需要替换明细，可以先 tx.Where("order_id = ?", order.ID).Delete(&sales.OrderItem{})，再批量创建 order.Items
		return nil
	})
}
