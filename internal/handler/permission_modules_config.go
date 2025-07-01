package handler

import (
	"djj-inventory-system/internal/model/rbac"
)

// PermissionModule 定义了一个权限模块的中英文名称、图标、描述和所有权限
// Key 用于代码中匹配（如 grantGroup），ModuleZh/ModuleEn 用于展示
var PermissionModules = []struct {
	Key           string
	ModuleZh      string
	ModuleEn      string
	Icon          string
	DescriptionZh string
	DescriptionEn string
	Permissions   []rbac.Permission
}{
	{
		Key:           "inventory",
		ModuleZh:      "库存管理",
		ModuleEn:      "Inventory Management",
		Icon:          "📦",
		DescriptionZh: "管理产品库存、入库出库等操作",
		DescriptionEn: "Manage product inventory, stock in and out operations",
		Permissions: []rbac.Permission{
			{ID: 101, Name: "inventory.view", Label: "查看库存", Description: "查看所有库存信息和统计数据"},
			{ID: 102, Name: "inventory.in", Label: "入库操作", Description: "执行商品入库操作"},
			{ID: 103, Name: "inventory.out", Label: "出库操作", Description: "执行商品出库操作"},
			{ID: 104, Name: "inventory.adjust", Label: "库存调整", Description: "调整库存数量和状态"},
			{ID: 105, Name: "inventory.transfer", Label: "库存转移", Description: "在不同仓库间转移库存"},
		},
	},
	{
		Key:           "sales",
		ModuleZh:      "销售管理",
		ModuleEn:      "Sales Management",
		Icon:          "💰",
		DescriptionZh: "管理销售订单、客户关系等",
		DescriptionEn: "Manage sales orders and customer relationships",
		Permissions: []rbac.Permission{
			{ID: 201, Name: "sales.view", Label: "查看销售", Description: "查看销售数据和报表"},
			{ID: 202, Name: "sales.create", Label: "新建销售订单", Description: "创建新的销售订单"},
			{ID: 203, Name: "sales.edit", Label: "编辑销售订单", Description: "修改现有销售订单"},
			{ID: 204, Name: "sales.delete", Label: "删除销售订单", Description: "删除销售订单"},
			{ID: 205, Name: "sales.approve", Label: "审批销售订单", Description: "审批销售订单"},
		},
	},
	{
		Key:           "quote",
		ModuleZh:      "报价管理",
		ModuleEn:      "Quote Management",
		Icon:          "📋",
		DescriptionZh: "管理客户报价和审批流程",
		DescriptionEn: "Manage customer quotes and approval workflows",
		Permissions: []rbac.Permission{
			{ID: 301, Name: "quote.view", Label: "查看报价", Description: "查看所有报价信息"},
			{ID: 302, Name: "quote.create", Label: "创建报价", Description: "为客户创建新报价"},
			{ID: 303, Name: "quote.edit", Label: "编辑报价", Description: "修改现有报价"},
			{ID: 304, Name: "quote.approve", Label: "审批报价", Description: "审批客户报价"},
			{ID: 305, Name: "quote.reject", Label: "拒绝报价", Description: "拒绝客户报价"},
		},
	},
	{
		Key:           "finance",
		ModuleZh:      "财务管理",
		ModuleEn:      "Finance Management",
		Icon:          "💳",
		DescriptionZh: "管理财务数据、账单和支付",
		DescriptionEn: "Manage financial data, invoicing and payments",
		Permissions: []rbac.Permission{
			{ID: 401, Name: "finance.view", Label: "查看财务", Description: "查看财务报表和数据"},
			{ID: 402, Name: "finance.invoice", Label: "开具发票", Description: "为客户开具发票"},
			{ID: 403, Name: "finance.payment", Label: "处理付款", Description: "处理客户付款"},
			{ID: 404, Name: "finance.refund", Label: "处理退款", Description: "处理客户退款"},
		},
	},
	{
		Key:           "user",
		ModuleZh:      "用户管理",
		ModuleEn:      "User Management",
		Icon:          "👥",
		DescriptionZh: "管理系统用户和权限",
		DescriptionEn: "Manage system users and permissions",
		Permissions: []rbac.Permission{
			{ID: 501, Name: "user.view", Label: "查看用户", Description: "查看系统用户列表"},
			{ID: 502, Name: "user.create", Label: "创建用户", Description: "创建新的系统用户"},
			{ID: 503, Name: "user.edit", Label: "编辑用户", Description: "修改用户信息"},
			{ID: 504, Name: "user.delete", Label: "删除用户", Description: "删除系统用户"},
			{ID: 505, Name: "user.permission", Label: "管理权限", Description: "管理用户权限配置"},
		},
	},
	{
		Key:           "system",
		ModuleZh:      "系统设置",
		ModuleEn:      "System Settings",
		Icon:          "⚙️",
		DescriptionZh: "系统配置和参数设置",
		DescriptionEn: "Configure system parameters and settings",
		Permissions: []rbac.Permission{
			{ID: 601, Name: "system.config", Label: "系统配置", Description: "修改系统配置参数"},
			{ID: 602, Name: "system.backup", Label: "数据备份", Description: "执行数据备份操作"},
			{ID: 603, Name: "system.restore", Label: "数据恢复", Description: "执行数据恢复操作"},
			{ID: 604, Name: "system.log", Label: "查看日志", Description: "查看系统操作日志"},
		},
	},
}
