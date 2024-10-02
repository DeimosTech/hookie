package hook

import (
	"context"
)

// AuditLogModels Registry for audit-log-enabled models
var AuditLogModels = make(map[string]bool)

func init() {
	err := WatchAndInjectHooks(context.Background(), ".")
	if err != nil {
		panic(err)
	}
}

// RegisterModel Register the model for audit logging
func RegisterModel(key string) {
	AuditLogModels[key] = true
}
