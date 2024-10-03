package hooks

import (
	"context"
	"fmt"
	"github.com/DeimosTech/hookie/db/mongo"
	in "github.com/DeimosTech/hookie/instance"
	"github.com/DeimosTech/hookie/internal/hook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"reflect"
	"time"
)

type DefaultHooks struct {
	l *slog.Logger
}

func NewDefaultHook() *DefaultHooks {
	return &DefaultHooks{l: slog.Default()}
}

func (h *DefaultHooks) PreSave(ctx context.Context, model interface{}, filter interface{}, col, ops, docId string) {
	// Trigger PreSave hook if defined by user, else run default
	if hasPreSaveHook(model) {
		model.(in.Hook).PreSave(ctx, model, filter, col, ops, docId)
		return
	}
	h.l.Info("default PreSave hook triggered")
}

func (h *DefaultHooks) PostSave(ctx context.Context, model interface{}, filter interface{}, col, ops, docId string) {
	if hasPostSaveHook(model) {
		model.(in.Hook).PostSave(ctx, model, filter, col, ops, docId)
		return
	}
	if isAuditLogEnabled(model) {
		db := mongo.GetDbConnection()
		if ops == "insert" {
			currentTime := time.Now()
			state, err := structToMap(model)
			if err != nil {
				h.l.Error(err.Error())
				return
			}
			auditLogMeta := in.AuditLogMeta{
				Id:                   primitive.NewObjectID(),
				AuditEvent:           ops,                      // Insert, Update, etc.
				AuditURL:             "example.com",            // Populate based on your logic
				AuditIPAddress:       "127.0.0.1",              // This can come from your request metadata
				AuditUserAgent:       "Mozilla/5.0",            // Populate from request metadata
				AuditTags:            []string{"audit", "log"}, // Customize tags
				AuditCreatedAt:       &currentTime,
				UserID:               "user_id",   // Set current user ID
				UserType:             "user_type", // Set current user type
				DocumentCurrentState: state,
			}
			_, err = db.Database.Collection("audit_logs_meta").InsertOne(context.Background(), auditLogMeta)
			if err != nil {
				db.Logger.Error(err.Error())
				return
			}

		} else if ops == "update" {
			auditFilter := bson.D{{"document_current_state._id", docId}}
			findOneOpts := options.FindOne()
			var auditLogMeta in.AuditLogMeta
			err := db.Database.Collection("audit_logs_meta").FindOne(ctx, auditFilter, findOneOpts).Decode(&auditLogMeta)
			if err != nil {
				h.l.Error(err.Error())
				return
			}
		}
	}
	h.l.Info("default PostSave hook triggered")
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

// hasPreSaveHook Check if the model has a PreSave method (custom user hook)
func hasPreSaveHook(model interface{}) bool {
	_, ok := reflect.TypeOf(model).MethodByName("PreSave")
	return ok
}

// hasPostSaveHook Check if the model has a PostSave method (custom user hook)
func hasPostSaveHook(model interface{}) bool {
	_, ok := reflect.TypeOf(model).MethodByName("PostSave")
	return ok
}

// SaveAuditLog Function to save an audit log after saving the model
func saveAuditLog(model interface{}) {
	fmt.Printf("Audit log saved for model: %T\n", model)
}

func structToMap(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)

	// Check if the input is a pointer and get the element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ensure the value is a struct
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", v.Kind())
	}

	// Iterate over the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)

		// You can also handle the JSON tag here if needed
		jsonTag := field.Tag.Get("json")

		// If the JSON tag is present, use it as the key
		if jsonTag != "" {
			result[jsonTag] = value.Interface()
		} else {
			result[field.Name] = value.Interface()
		}
	}

	return result, nil
}
