// internal/model/audit_types.go
package common

// AuditedTable 是被审计的表名
type AuditedTable string

// AuditAction 是审计动作类型
type AuditAction string

const (
	ActionCreate AuditAction = "CREATE"
	ActionUpdate AuditAction = "UPDATE"
	ActionDelete AuditAction = "DELETE"
)
