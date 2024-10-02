package hook

import (
	"context"
	"reflect"
	"sync"
)

// AuditLogModels Registry for audit-log-enabled models
var AuditLogModels sync.Map

func init() {
	err := WatchAndInjectHooks(context.Background(), ".")
	if err != nil {
		panic(err)
	}
}

// RegisterModel Register the model for audit logging
func RegisterModel(model interface{}) {
	modelType := reflect.TypeOf(model).Elem() // Get the model type
	AuditLogModels.Store(modelType, true)
}
