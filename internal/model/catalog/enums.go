// internal/model/catalog/enums.go
package catalog

// ProductStatus 对应 postgresql 的 product_status_enum
type ProductStatus string

const (
	StatusDraft           ProductStatus = "draft"
	StatusPendingTech     ProductStatus = "pending_tech"
	StatusPendingPurchase ProductStatus = "pending_purchase"
	StatusPendingFinance  ProductStatus = "pending_finance"
	StatusReadyPublished  ProductStatus = "ready_published"
	StatusPublished       ProductStatus = "published"
	StatusRejected        ProductStatus = "rejected"
	StatusClosed          ProductStatus = "closed"
)

// ApplicationStatus 对应 application_status_enum
type ApplicationStatus string

const (
	AppOpen   ApplicationStatus = "open"
	AppClosed ApplicationStatus = "closed"
)

// ProductType 对应 product_type_enum
type ProductType string

const (
	TypeMachine    ProductType = "machine"
	TypeParts      ProductType = "parts"
	TypeAttachment ProductType = "attachment"
	TypeTools      ProductType = "tools"
	TypeOthers     ProductType = "others"
)

// 如果你的 category/subcategory/tertiary 也走固定枚举：
type Category string

const (
	CategoryMachine     Category = "Machine"
	CategoryParts       Category = "Parts"
	CategoryTools       Category = "Tools"
	CategoryAccessories Category = "Accessories"
)

// Subcategory／TertiaryCategory 可以类似定义
type Subcategory string
type TertiaryCategory string
