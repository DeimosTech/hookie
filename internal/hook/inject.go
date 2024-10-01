package hook

import (
	"fmt"
	"log/slog"
	"reflect"
	"sync"
)

// Registry for audit-log-enabled models
var auditLogModels sync.Map

// Hook interface for custom hooks
type Hook interface {
	BeforeInsert()
	AfterInsert()
}

type Inject struct {
}

// Step 2: Automatic detection of models that embed BaseModel or have in.In fields
func init() {
	logger := slog.Default()

	// Example models that would be registered
	var models []interface{}

	for _, model := range models {
		// Register models that embed BaseModel or in.In
		if HasBaseModelOrInField(model) {
			RegisterModel(model)
			logger.Info("Registered model for audit logging: %T\n", model)
		}
	}
}

// HasBaseModelOrInField Function to check if a model contains BaseModel or an audit marker (like hook.Inject)
func HasBaseModelOrInField(model interface{}) bool {
	typ := reflect.TypeOf(model).Elem() // We assume it's a pointer
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Check if the field is of type BaseModel or in.In
		if field.Type == reflect.TypeOf(Inject{}) {
			return true
		}
	}
	return false
}

// RegisterModel Register the model for audit logging
func RegisterModel(model interface{}) {
	modelType := reflect.TypeOf(model).Elem() // Get the model type
	auditLogModels.Store(modelType, true)
}

func DefaultBeforeInsert(model interface{}) {
	// Trigger BeforeInsert hook if defined by user, else run default
	if hasBeforeInsertHook(model) {
		model.(Hook).BeforeInsert()
	}
	if isAuditLogEnabled(model) {
		saveAuditLog(model)
	}
}

func DefaultAfterInsert(model interface{}) {
	if hasAfterInsertHook(model) {
		model.(Hook).AfterInsert()
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
