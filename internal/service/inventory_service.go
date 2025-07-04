// ===========================================
// 2. 库存Service层
// ===========================================

// internal/service/inventory_service.go
package service

import (
	"context"
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/inventory"
	"djj-inventory-system/internal/repository"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type InventoryService struct {
	repo   *repository.InventoryRepository
	logger *zap.Logger
}

func NewInventoryService(repo *repository.InventoryRepository, logger *zap.Logger) *InventoryService {
	return &InventoryService{
		repo:   repo,
		logger: logger,
	}
}

// ==== 库存查询相关 ====

// GetProductStock 获取产品的库存信息
func (s *InventoryService) GetProductStock(ctx context.Context, productID uint) ([]catalog.ProductStock, error) {
	stocks, err := s.repo.GetStockByProductID(ctx, productID)
	if err != nil {
		s.logger.Error("Failed to get product stock", zap.Uint("productID", productID), zap.Error(err))
		return nil, fmt.Errorf("failed to get product stock: %w", err)
	}

	return stocks, nil
}

// GetProductStockInWarehouse 获取产品在特定仓库的库存
func (s *InventoryService) GetProductStockInWarehouse(ctx context.Context, productID, warehouseID uint) (*catalog.ProductStock, error) {
	stock, err := s.repo.GetStockByProductAndWarehouse(ctx, productID, warehouseID)
	if err != nil {
		s.logger.Error("Failed to get product stock in warehouse",
			zap.Uint("productID", productID),
			zap.Uint("warehouseID", warehouseID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get product stock in warehouse: %w", err)
	}

	return stock, nil
}

// ==== 库存操作相关 ====

// StockIn 入库操作
func (s *InventoryService) StockIn(ctx context.Context, productID, warehouseID uint, quantity int, operator, note string) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if operator == "" {
		return errors.New("operator is required")
	}

	err := s.repo.ProcessStockMovement(ctx, productID, warehouseID, quantity,
		inventory.TransactionTypeIn, operator, note)
	if err != nil {
		s.logger.Error("Failed to process stock in",
			zap.Uint("productID", productID),
			zap.Uint("warehouseID", warehouseID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to process stock in: %w", err)
	}

	s.logger.Info("Stock in processed successfully",
		zap.Uint("productID", productID),
		zap.Uint("warehouseID", warehouseID),
		zap.Int("quantity", quantity),
		zap.String("operator", operator))

	return nil
}

// StockOut 出库操作
func (s *InventoryService) StockOut(ctx context.Context, productID, warehouseID uint, quantity int, operator, note string) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if operator == "" {
		return errors.New("operator is required")
	}

	err := s.repo.ProcessStockMovement(ctx, productID, warehouseID, quantity,
		inventory.TransactionTypeOut, operator, note)
	if err != nil {
		s.logger.Error("Failed to process stock out",
			zap.Uint("productID", productID),
			zap.Uint("warehouseID", warehouseID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to process stock out: %w", err)
	}

	s.logger.Info("Stock out processed successfully",
		zap.Uint("productID", productID),
		zap.Uint("warehouseID", warehouseID),
		zap.Int("quantity", quantity),
		zap.String("operator", operator))

	return nil
}

// Sale 销售操作
func (s *InventoryService) Sale(ctx context.Context, productID, warehouseID uint, quantity int, operator, note string) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if operator == "" {
		return errors.New("operator is required")
	}

	err := s.repo.ProcessStockMovement(ctx, productID, warehouseID, quantity,
		inventory.TransactionTypeSale, operator, note)
	if err != nil {
		s.logger.Error("Failed to process sale",
			zap.Uint("productID", productID),
			zap.Uint("warehouseID", warehouseID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to process sale: %w", err)
	}

	s.logger.Info("Sale processed successfully",
		zap.Uint("productID", productID),
		zap.Uint("warehouseID", warehouseID),
		zap.Int("quantity", quantity),
		zap.String("operator", operator))

	return nil
}

// GetProductTransactions 获取产品的所有事务记录
func (s *InventoryService) GetProductTransactions(ctx context.Context, productID uint, offset, limit int) ([]inventory.InventoryTransaction, int64, error) {
	transactions, total, err := s.repo.GetTransactionsByProduct(ctx, productID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get product transactions",
			zap.Uint("productID", productID),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get product transactions: %w", err)
	}

	return transactions, total, nil
}

// BatchStockUpdate 批量更新库存
func (s *InventoryService) BatchStockUpdate(ctx context.Context, updates []StockUpdateRequest, operator string) error {
	if operator == "" {
		return errors.New("operator is required")
	}

	for i, update := range updates {
		if err := s.validateStockUpdate(update); err != nil {
			return fmt.Errorf("validation failed for update %d: %w", i, err)
		}
	}

	// 处理每个更新
	for _, update := range updates {
		var err error
		switch update.Type {
		case "IN":
			err = s.StockIn(ctx, update.ProductID, update.WarehouseID, update.Quantity, operator, update.Note)
		case "OUT":
			err = s.StockOut(ctx, update.ProductID, update.WarehouseID, update.Quantity, operator, update.Note)
		case "SALE":
			err = s.Sale(ctx, update.ProductID, update.WarehouseID, update.Quantity, operator, update.Note)
		default:
			err = fmt.Errorf("unknown update type: %s", update.Type)
		}

		if err != nil {
			s.logger.Error("Failed to process batch stock update",
				zap.Uint("productID", update.ProductID),
				zap.Uint("warehouseID", update.WarehouseID),
				zap.String("type", update.Type),
				zap.Error(err))
			return fmt.Errorf("failed to process update for product %d: %w", update.ProductID, err)
		}
	}

	s.logger.Info("Batch stock update completed successfully",
		zap.Int("updateCount", len(updates)),
		zap.String("operator", operator))

	return nil
}

func (s *InventoryService) validateStockUpdate(update StockUpdateRequest) error {
	if update.ProductID == 0 {
		return errors.New("product ID is required")
	}
	if update.WarehouseID == 0 {
		return errors.New("warehouse ID is required")
	}
	if update.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if update.Type != "IN" && update.Type != "OUT" && update.Type != "SALE" {
		return errors.New("invalid update type")
	}
	return nil
}

// StockUpdateRequest 批量更新请求
type StockUpdateRequest struct {
	ProductID   uint   `json:"productId" validate:"required"`
	WarehouseID uint   `json:"warehouseId" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Type        string `json:"type" validate:"required,oneof=IN OUT SALE"`
	Note        string `json:"note"`
}
