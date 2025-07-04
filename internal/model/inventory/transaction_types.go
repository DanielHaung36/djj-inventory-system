package inventory

// TransactionType 扩展的事务类型
type TransactionType string

const (
	// 基本库存操作
	TransactionTypeIn   TransactionType = "IN"   // 入库
	TransactionTypeOut  TransactionType = "OUT"  // 出库
	TransactionTypeSale TransactionType = "SALE" // 销售

	// 预留相关操作
	TransactionTypeReserve TransactionType = "RESERVE" // 预留库存
	TransactionTypeRelease TransactionType = "RELEASE" // 释放预留

	// 调整操作
	TransactionTypeAdjust      TransactionType = "ADJUST"       // 库存调整
	TransactionTypeTransferIn  TransactionType = "TRANSFER_IN"  // 转入
	TransactionTypeTransferOut TransactionType = "TRANSFER_OUT" // 转出

	// 其他操作
	TransactionTypeReturn  TransactionType = "RETURN"  // 退货
	TransactionTypeDamage  TransactionType = "DAMAGE"  // 损坏
	TransactionTypeExpired TransactionType = "EXPIRED" // 过期
	TransactionTypeStolen  TransactionType = "STOLEN"  // 丢失/被盗
)

// TransactionTypeInfo 事务类型信息
type TransactionTypeInfo struct {
	Type        TransactionType `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Impact      string          `json:"impact"` // "positive", "negative", "neutral"
}

// GetTransactionTypeInfo 获取事务类型信息映射表
func GetTransactionTypeInfo() map[TransactionType]TransactionTypeInfo {
	return map[TransactionType]TransactionTypeInfo{
		TransactionTypeIn: {
			Type:        TransactionTypeIn,
			Name:        "入库",
			Description: "商品入库操作，增加库存",
			Impact:      "positive",
		},
		TransactionTypeOut: {
			Type:        TransactionTypeOut,
			Name:        "出库",
			Description: "商品出库操作，减少库存",
			Impact:      "negative",
		},
		TransactionTypeSale: {
			Type:        TransactionTypeSale,
			Name:        "销售",
			Description: "商品销售操作，减少库存",
			Impact:      "negative",
		},
		TransactionTypeReserve: {
			Type:        TransactionTypeReserve,
			Name:        "预留",
			Description: "预留库存，不影响实际库存但影响可用库存",
			Impact:      "neutral",
		},
		TransactionTypeRelease: {
			Type:        TransactionTypeRelease,
			Name:        "释放",
			Description: "释放预留库存，增加可用库存",
			Impact:      "neutral",
		},
		TransactionTypeAdjust: {
			Type:        TransactionTypeAdjust,
			Name:        "调整",
			Description: "库存数量调整，可正可负",
			Impact:      "neutral",
		},
		TransactionTypeTransferIn: {
			Type:        TransactionTypeTransferIn,
			Name:        "转入",
			Description: "从其他仓库转入，增加库存",
			Impact:      "positive",
		},
		TransactionTypeTransferOut: {
			Type:        TransactionTypeTransferOut,
			Name:        "转出",
			Description: "转出到其他仓库，减少库存",
			Impact:      "negative",
		},
		TransactionTypeReturn: {
			Type:        TransactionTypeReturn,
			Name:        "退货",
			Description: "客户退货，增加库存",
			Impact:      "positive",
		},
		TransactionTypeDamage: {
			Type:        TransactionTypeDamage,
			Name:        "损坏",
			Description: "商品损坏，减少库存",
			Impact:      "negative",
		},
		TransactionTypeExpired: {
			Type:        TransactionTypeExpired,
			Name:        "过期",
			Description: "商品过期，减少库存",
			Impact:      "negative",
		},
		TransactionTypeStolen: {
			Type:        TransactionTypeStolen,
			Name:        "丢失",
			Description: "商品丢失或被盗，减少库存",
			Impact:      "negative",
		},
	}
}

// IsValidTransactionType 检查是否为有效的事务类型
func IsValidTransactionType(txType TransactionType) bool {
	_, exists := GetTransactionTypeInfo()[txType]
	return exists
}

// GetImpactDirection 获取事务类型对库存的影响方向
// 返回值：1 表示增加，-1 表示减少，0 表示不影响实际库存
func GetImpactDirection(txType TransactionType) int {
	switch txType {
	case TransactionTypeIn, TransactionTypeTransferIn, TransactionTypeReturn:
		return 1
	case TransactionTypeOut, TransactionTypeSale, TransactionTypeTransferOut,
		TransactionTypeDamage, TransactionTypeExpired, TransactionTypeStolen:
		return -1
	case TransactionTypeReserve, TransactionTypeRelease, TransactionTypeAdjust:
		return 0
	default:
		return 0
	}
}
