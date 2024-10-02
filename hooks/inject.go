package hooks

import (
	"context"
	"fmt"
	"github.com/DeimosTech/hookie/db/mongo"
	in "github.com/DeimosTech/hookie/instance"
	"github.com/DeimosTech/hookie/internal/hook"
	"reflect"
)

type DefaultHooks struct {
}

func (h *DefaultHooks) BeforeInsert(model interface{}) {
	// Trigger BeforeInsert hook if defined by user, else run default
	if hasBeforeInsertHook(model) {
		model.(in.Hook).BeforeInsert(model)
		return
	}
	fmt.Println("Default BeforeInsert hook called")
}

func (h *DefaultHooks) AfterInsert(model interface{}) {
	if hasAfterInsertHook(model) {
		model.(in.Hook).AfterInsert(model)
		return
	}
	if isAuditLogEnabled(model) {
		db := mongo.GetDbConnection()
		_, err := db.Database.Collection("audit_logs").InsertOne(context.Background(), model)
		if err != nil {
			db.Logger.Error(err.Error())
		}
	}
	fmt.Println("Default AfterInsert hook called")
}

// isAuditLogEnabled Function to check if the model has audit logging enabled
func isAuditLogEnabled(model interface{}) bool {
	modelType := reflect.TypeOf(model)

	// If the modelType is a pointer, get the underlying type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Ensure it's a struct
	if modelType.Kind() != reflect.Struct {
		return false
	}

	// Get the type's name
	typeName := modelType.Name()

	// Get the package path using the types package
	pkgPath := modelType.PkgPath()
	return hook.AuditLogModels[pkgPath+typeName]
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
