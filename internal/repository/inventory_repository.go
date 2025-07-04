// internal/repository/inventory_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/inventory"

	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// ==== ProductStock 相关操作 ====

// GetStockByProductID 根据产品ID获取所有仓库的库存
func (r *InventoryRepository) GetStockByProductID(ctx context.Context, productID uint) ([]catalog.ProductStock, error) {
	var stocks []catalog.ProductStock
	err := r.db.WithContext(ctx).
		Preload("Warehouse"). // 预加载仓库信息
		Preload("Product"). // 预加载产品信息
		Where("product_id = ?", productID).
		Find(&stocks).Error
	return stocks, err
}

// GetStockByProductAndWarehouse 根据产品ID和仓库ID获取特定库存
func (r *InventoryRepository) GetStockByProductAndWarehouse(ctx context.Context, productID, warehouseID uint) (*catalog.ProductStock, error) {
	var stock catalog.ProductStock
	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product").
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&stock).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &stock, err
}

// GetStocksByWarehouse 根据仓库ID获取所有库存（支持分页）
func (r *InventoryRepository) GetStocksByWarehouse(ctx context.Context, warehouseID uint, offset, limit int) ([]catalog.ProductStock, int64, error) {
	var stocks []catalog.ProductStock
	var total int64

	query := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Warehouse").
		Where("warehouse_id = ?", warehouseID)

	// 计算总数
	if err := query.Model(&catalog.ProductStock{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Find(&stocks).Error
	return stocks, total, err
}

// GetLowStockProducts 获取低库存产品（现有量 <= 最小库存量）
func (r *InventoryRepository) GetLowStockProducts(ctx context.Context, threshold int) ([]catalog.ProductStock, error) {
	var stocks []catalog.ProductStock
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Warehouse").
		Where("on_hand <= ?", threshold).
		Find(&stocks).Error
	return stocks, err
}

// CreateOrUpdateStock 创建或更新库存记录
func (r *InventoryRepository) CreateOrUpdateStock(ctx context.Context, stock *catalog.ProductStock) error {
	// 使用 GORM 的 Save 方法，如果存在则更新，不存在则创建
	return r.db.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ?", stock.ProductID, stock.WarehouseID).
		Assign(map[string]interface{}{
			"on_hand":    stock.OnHand,
			"reserved":   stock.Reserved,
			"updated_at": time.Now(),
		}).
		FirstOrCreate(stock).Error
}

// UpdateStockQuantity 更新库存数量（原子操作，支持事务）
func (r *InventoryRepository) UpdateStockQuantity(ctx context.Context, productID, warehouseID uint, onHandDelta, reservedDelta int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先检查当前库存
		var stock catalog.ProductStock
		if err := tx.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
			First(&stock).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 如果不存在，创建新记录
				stock = catalog.ProductStock{
					ProductID:   productID,
					WarehouseID: warehouseID,
					OnHand:      0,
					Reserved:    0,
				}
			} else {
				return err
			}
		}

		// 计算新的库存量
		newOnHand := stock.OnHand + onHandDelta
		newReserved := stock.Reserved + reservedDelta

		// 检查库存不能为负
		if newOnHand < 0 {
			return errors.New("insufficient stock: on_hand would be negative")
		}
		if newReserved < 0 {
			return errors.New("insufficient stock: reserved would be negative")
		}

		// 更新库存
		return tx.Model(&stock).Updates(map[string]interface{}{
			"on_hand":    newOnHand,
			"reserved":   newReserved,
			"updated_at": time.Now(),
		}).Error
	})
}

// DeleteStock 删除库存记录
func (r *InventoryRepository) DeleteStock(ctx context.Context, productID, warehouseID uint) error {
	return r.db.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Delete(&catalog.ProductStock{}).Error
}

// ==== InventoryTransaction 相关操作 ====

// CreateTransaction 创建库存事务记录
func (r *InventoryRepository) CreateTransaction(ctx context.Context, tx *inventory.InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetTransactionsByInventory 根据库存ID获取事务记录
func (r *InventoryRepository) GetTransactionsByInventory(ctx context.Context, inventoryID uint, offset, limit int) ([]inventory.InventoryTransaction, int64, error) {
	var transactions []inventory.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).
		Preload("Inventory").
		Preload("Inventory.Product").
		Preload("Inventory.Warehouse").
		Where("inventory_id = ?", inventoryID)

	// 计算总数
	if err := query.Model(&inventory.InventoryTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按时间倒序
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	return transactions, total, err
}

// GetTransactionsByProduct 根据产品ID获取所有相关事务记录
func (r *InventoryRepository) GetTransactionsByProduct(ctx context.Context, productID uint, offset, limit int) ([]inventory.InventoryTransaction, int64, error) {
	var transactions []inventory.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).
		Preload("Inventory").
		Preload("Inventory.Product").
		Preload("Inventory.Warehouse").
		Joins("JOIN product_stocks ON inventory_transaction.inventory_id = product_stocks.id").
		Where("product_stocks.product_id = ?", productID)

	// 计算总数
	if err := query.Model(&inventory.InventoryTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	return transactions, total, err
}

// GetTransactionsByWarehouse 根据仓库ID获取所有相关事务记录
func (r *InventoryRepository) GetTransactionsByWarehouse(ctx context.Context, warehouseID uint, offset, limit int) ([]inventory.InventoryTransaction, int64, error) {
	var transactions []inventory.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).
		Preload("Inventory").
		Preload("Inventory.Product").
		Preload("Inventory.Warehouse").
		Joins("JOIN product_stocks ON inventory_transaction.inventory_id = product_stocks.id").
		Where("product_stocks.warehouse_id = ?", warehouseID)

	// 计算总数
	if err := query.Model(&inventory.InventoryTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	return transactions, total, err
}

// GetTransactionsByDateRange 根据时间范围获取事务记录
func (r *InventoryRepository) GetTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time, offset, limit int) ([]inventory.InventoryTransaction, int64, error) {
	var transactions []inventory.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).
		Preload("Inventory").
		Preload("Inventory.Product").
		Preload("Inventory.Warehouse").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	// 计算总数
	if err := query.Model(&inventory.InventoryTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	return transactions, total, err
}

// ==== 复合操作 ====

// ProcessStockMovement 处理库存移动（带事务记录）
func (r *InventoryRepository) ProcessStockMovement(ctx context.Context, productID, warehouseID uint, quantity int, txType inventory.TransactionType, operator, note string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 更新库存
		var onHandDelta, reservedDelta int
		switch txType {
		case inventory.TransactionTypeIn:
			onHandDelta = quantity
		case inventory.TransactionTypeOut, inventory.TransactionTypeSale:
			onHandDelta = -quantity
		default:
			return errors.New("invalid transaction type")
		}

		// 检查并更新库存
		var stock catalog.ProductStock
		if err := tx.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
			First(&stock).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新库存记录
				stock = catalog.ProductStock{
					ProductID:   productID,
					WarehouseID: warehouseID,
					OnHand:      0,
					Reserved:    0,
				}
				if err := tx.Create(&stock).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 检查库存是否足够
		if stock.OnHand+onHandDelta < 0 {
			return errors.New("insufficient stock")
		}

		// 更新库存
		if err := tx.Model(&stock).Updates(map[string]interface{}{
			"on_hand":    stock.OnHand + onHandDelta,
			"reserved":   stock.Reserved + reservedDelta,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// 2. 创建事务记录
		transaction := &inventory.InventoryTransaction{
			InventoryID: stock.ID,
			TxType:      txType,
			Quantity:    quantity,
			Operator:    operator,
			Note:        note,
			CreatedAt:   time.Now(),
		}

		return tx.Create(transaction).Error
	})
}

// ReserveStock 预留库存
func (r *InventoryRepository) ReserveStock(ctx context.Context, productID, warehouseID uint, quantity int, operator, note string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取当前库存
		var stock catalog.ProductStock
		if err := tx.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
			First(&stock).Error; err != nil {
			return err
		}

		// 检查可用库存是否足够
		available := stock.OnHand - stock.Reserved
		if available < quantity {
			return errors.New("insufficient available stock for reservation")
		}

		// 更新预留量
		if err := tx.Model(&stock).Updates(map[string]interface{}{
			"reserved":   stock.Reserved + quantity,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// 创建事务记录
		transaction := &inventory.InventoryTransaction{
			InventoryID: stock.ID,
			TxType:      inventory.TransactionTypeReserve,
			Quantity:    quantity,
			Operator:    operator,
			Note:        note,
			CreatedAt:   time.Now(),
		}

		return tx.Create(transaction).Error
	})
}

// ReleaseReservedStock 释放预留库存
func (r *InventoryRepository) ReleaseReservedStock(ctx context.Context, productID, warehouseID uint, quantity int, operator, note string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取当前库存
		var stock catalog.ProductStock
		if err := tx.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
			First(&stock).Error; err != nil {
			return err
		}

		// 检查预留量是否足够
		if stock.Reserved < quantity {
			return errors.New("insufficient reserved stock to release")
		}

		// 更新预留量
		if err := tx.Model(&stock).Updates(map[string]interface{}{
			"reserved":   stock.Reserved - quantity,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// 创建事务记录
		transaction := &inventory.InventoryTransaction{
			InventoryID: stock.ID,
			TxType:      inventory.TransactionTypeRelease,
			Quantity:    quantity,
			Operator:    operator,
			Note:        note,
			CreatedAt:   time.Now(),
		}

		return tx.Create(transaction).Error
	})
}

// GetInventorySummary 获取库存汇总信息
func (r *InventoryRepository) GetInventorySummary(ctx context.Context, productID uint) (map[string]interface{}, error) {
	var result struct {
		TotalOnHand    int `json:"total_on_hand"`
		TotalReserved  int `json:"total_reserved"`
		TotalAvailable int `json:"total_available"`
		WarehouseCount int `json:"warehouse_count"`
	}

	err := r.db.WithContext(ctx).
		Model(&catalog.ProductStock{}).
		Select(`
			COALESCE(SUM(on_hand), 0) as total_on_hand,
			COALESCE(SUM(reserved), 0) as total_reserved,
			COALESCE(SUM(on_hand - reserved), 0) as total_available,
			COUNT(*) as warehouse_count
		`).
		Where("product_id = ?", productID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalOnHand":    result.TotalOnHand,
		"totalReserved":  result.TotalReserved,
		"totalAvailable": result.TotalAvailable,
		"warehouseCount": result.WarehouseCount,
	}, nil
}
