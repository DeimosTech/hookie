package hook

import (
	"context"
	"fmt"
	in "github.com/DeimosTech/hookie/instance"
	"reflect"
	"sync"
)

// Registry for audit-log-enabled models
var auditLogModels sync.Map

func init() {
	err := WatchAndInjectHooks(context.Background(), ".")
	if err != nil {
		panic(err)
	}
}

// RegisterModel Register the model for audit logging
func RegisterModel(model interface{}) {
	modelType := reflect.TypeOf(model).Elem() // Get the model type
	auditLogModels.Store(modelType, true)
}

func DefaultBeforeInsert(model interface{}) {
	// Trigger BeforeInsert hook if defined by user, else run default
	if hasBeforeInsertHook(model) {
		model.(in.Hook).BeforeInsert()
	}
	if isAuditLogEnabled(model) {
		saveAuditLog(model)
	}
}

func DefaultAfterInsert(model interface{}) {
	if hasAfterInsertHook(model) {
		model.(in.Hook).AfterInsert()
	}
	fmt.Println("Default AfterInsert hook called")
}

// isAuditLogEnabled Function to check if the model has audit logging enabled
func isAuditLogEnabled(model interface{}) bool {
	modelType := reflect.TypeOf(model).Elem() // Get the model type
	_, exists := auditLogModels.Load(modelType)
	return exists
}

// hasBeforeInsertHook Check if the model has a BeforeInsert method (custom user hook)
func hasBeforeInsertHook(model interface{}) bool {
	_, ok := reflect.TypeOf(model).MethodByName("BeforeInsert")
	return ok
}

// hasAfterInsertHook Check if the model has a AfterInsert method (custom user hook)
func hasAfterInsertHook(model interface{}) bool {
	_, ok := reflect.TypeOf(model).MethodByName("AfterInsert")
	return ok
}

// SaveAuditLog Function to save an audit log after saving the model
func saveAuditLog(model interface{}) {
	fmt.Printf("Audit log saved for model: %T\n", model)
}
