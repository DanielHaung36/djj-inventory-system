package service

import (
	"context"
	"encoding/json"
	"time"

	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/dto"
	"djj-inventory-system/internal/repository"

	"gorm.io/datatypes"
)

type ProductService struct {
	ProdRepo  *repository.ProductRepository
	StockRepo *repository.StockRepository
}

func NewProductService(
	pr *repository.ProductRepository,
	sr *repository.StockRepository,
) *ProductService {
	return &ProductService{ProdRepo: pr, StockRepo: sr}
}

// Create 新建产品
func (s *ProductService) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {

	p := &catalog.Product{
		DJJCode:          req.DJJCode,
		Status:           catalog.ProductStatus(req.Status),
		Supplier:         req.Supplier,
		ManufacturerCode: req.ManufacturerCode,
		Category:         catalog.Category(req.Category),
		Subcategory:      req.Subcategory,
		TertiaryCategory: req.TertiaryCategory,
		NameCN:           req.NameCN,
		NameEN:           req.NameEN,
		Specs:            req.Specs,
		Standards:        req.Standards,
		Unit:             req.Unit,
		Currency:         req.Currency,
		RRPPrice:         req.RRPPrice,
		Price:            req.Price,
		StandardWarranty: req.StandardWarranty,
		Remarks:          req.Remarks,
		WeightKG:         req.WeightKG,
		LiftCapacityKG:   req.LiftCapacityKG,
		LiftHeightMM:     req.LiftHeightMM,
		PowerSource:      req.PowerSource,
		OtherSpecs:       datatypes.JSON(req.OtherSpecs),
		Warranty:         req.Warranty,
		MarketingInfo:    req.MarketingInfo,
		TrainingDocs:     req.TrainingDocs,
		ProductURL:       req.ProductURL,
		TechnicalSpecs:   datatypes.JSON(req.TechnicalSpecs),
		ExtraInfo:        datatypes.JSON(req.OtherInfo),
	}
	for _, img := range req.Images {
		p.Images = append(p.Images, catalog.ProductImage{
			URL:       img.URL,
			Alt:       img.Alt,
			IsPrimary: img.IsPrimary,
			CreatedAt: time.Now(),
		})
	}
	// 写库
	if err := s.ProdRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	// 同步库存
	if err := s.syncStocks(ctx, p.ID, req.Stocks); err != nil {
		return nil, err
	}
	// 返回 DTO
	return s.toDTO(ctx, p.ID)
}

// Update 修改产品
func (s *ProductService) Update(ctx context.Context, id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	p, err := s.ProdRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	// 更新字段
	p.DJJCode = req.DJJCode
	p.Status = catalog.ProductStatus(req.Status)
	p.Supplier = req.Supplier
	p.ManufacturerCode = req.ManufacturerCode
	p.Category = catalog.Category(req.Category)
	p.Subcategory = req.Subcategory
	p.TertiaryCategory = req.TertiaryCategory
	p.NameCN = req.NameCN
	p.NameEN = req.NameEN
	p.Specs = req.Specs
	p.Standards = req.Standards
	p.Unit = req.Unit
	p.Currency = req.Currency
	p.RRPPrice = req.RRPPrice
	p.Price = req.Price
	p.StandardWarranty = req.StandardWarranty
	p.Remarks = req.Remarks
	p.WeightKG = req.WeightKG
	p.LiftCapacityKG = req.LiftCapacityKG
	p.LiftHeightMM = req.LiftHeightMM
	p.PowerSource = req.PowerSource
	p.OtherSpecs = datatypes.JSON(req.OtherSpecs)
	p.Warranty = req.Warranty
	p.MarketingInfo = req.MarketingInfo
	p.TrainingDocs = req.TrainingDocs
	p.ProductURL = req.ProductURL
	p.TechnicalSpecs = datatypes.JSON(req.TechnicalSpecs)
	p.ExtraInfo = datatypes.JSON(req.OtherInfo)

	if err := s.ProdRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	if err := s.syncStocks(ctx, p.ID, req.Stocks); err != nil {
		return nil, err
	}
	return s.toDTO(ctx, p.ID)
}

// Delete 删除产品
func (s *ProductService) Delete(ctx context.Context, id uint) error {
	return s.ProdRepo.Delete(ctx, id)
}

// GetByID 读取一条
func (s *ProductService) GetByID(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	return s.toDTO(ctx, id)
}

// List 列表
func (s *ProductService) List(ctx context.Context, offset, limit int) ([]dto.ProductResponse, int64, error) {
	models, total, err := s.ProdRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ProductResponse, len(models))
	for i := range models {
		out[i] = mapProductToResponse(&models[i])
	}
	return out, total, nil
}

// syncStocks 先删后增
func (s *ProductService) syncStocks(ctx context.Context, pid uint, stocks []dto.StockEntry) error {
	if err := s.StockRepo.DeleteByProduct(ctx, pid); err != nil {
		return err
	}
	for _, e := range stocks {
		ps := &catalog.ProductStock{
			ProductID:   pid,
			WarehouseID: e.WarehouseID,
			OnHand:      e.OnHand,
			UpdatedAt:   time.Now(),
		}
		if err := s.StockRepo.Create(ctx, ps); err != nil {
			return err
		}
	}
	return nil
}

// toDTO 读取并转换
func (s *ProductService) toDTO(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	p, err := s.ProdRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	r := mapProductToResponse(p)
	return &r, nil
}

// mapProductToResponse Model->DTO
func mapProductToResponse(p *catalog.Product) dto.ProductResponse {
	stocks := make([]dto.StockEntry, len(p.Stocks))
	for i, ps := range p.Stocks {
		stocks[i] = dto.StockEntry{
			WarehouseID:   ps.WarehouseID,
			WarehouseName: ps.Warehouse.Name,
			//实际库存
			OnHand: ps.OnHand,
		}
	}
	images := make([]dto.ProductImageDTO, len(p.Images))
	for i, a := range p.Images {
		images[i] = dto.ProductImageDTO{
			ID:        a.ID,
			URL:       a.URL,
			Alt:       a.Alt,
			IsPrimary: a.IsPrimary,
		}
	}
	return dto.ProductResponse{
		ID:               p.ID,
		DJJCode:          p.DJJCode,
		Status:           string(p.Status),
		Supplier:         p.Supplier,
		ManufacturerCode: p.ManufacturerCode,
		Category:         string(p.Category),
		Subcategory:      p.Subcategory,
		TertiaryCategory: p.TertiaryCategory,
		NameCN:           p.NameCN,
		NameEN:           p.NameEN,
		Specs:            p.Specs,
		Standards:        p.Standards,
		Unit:             p.Unit,
		Currency:         p.Currency,
		RRPPrice:         p.RRPPrice,
		StandardWarranty: p.StandardWarranty,
		Remarks:          p.Remarks,
		WeightKG:         p.WeightKG,
		LiftCapacityKG:   p.LiftCapacityKG,
		LiftHeightMM:     p.LiftHeightMM,
		PowerSource:      p.PowerSource,
		OtherSpecs:       json.RawMessage(p.OtherSpecs),
		Warranty:         p.StandardWarranty,
		MarketingInfo:    p.MarketingInfo,
		TrainingDocs:     p.TrainingDocs,
		ProductURL:       p.ProductURL,
		Stocks:           stocks,
		LastUpdate:       p.UpdatedAt,
		LastModifiedBy:   "",
		MonthlySales:     0,
		TotalSales:       0,
		ProfitMargin:     0,
		TechnicalSpecs:   json.RawMessage(p.TechnicalSpecs),
		OtherInfo:        json.RawMessage(p.ExtraInfo),
		Images:           images,
		SalesData:        nil,
		Version:          p.Version,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// helper to convert JSONPointer to *string
func ptrStringToPtrString(j datatypes.JSON) *string {
	if len(j) == 0 {
		return nil
	}
	s := string(j)
	return &s
}
