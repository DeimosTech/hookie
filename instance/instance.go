package in

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Hook interface for custom hooks
type Hook interface {
	PreSave(ctx context.Context, model interface{}, filter interface{}, col, ops, docId string)
	PostSave(ctx context.Context, model interface{}, filter interface{}, col, ops, docId string)
}

type Inject struct {
}

type Test struct {
	*Inject
	Id   string
	Name string
}

type AuditLogMeta struct {
	Id                   primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	AuditEvent           string                 `json:"audit_event,omitempty" bson:"audit_event,omitempty"`
	AuditURL             string                 `json:"audit_url,omitempty" bson:"audit_url,omitempty"`
	AuditIPAddress       string                 `json:"audit_ip_address,omitempty" bson:"audit_ip_address,omitempty"`
	AuditUserAgent       string                 `json:"audit_user_agent,omitempty" bson:"audit_user_agent,omitempty"`
	AuditTags            []string               `json:"audit_tags,omitempty" bson:"audit_tags,omitempty"`
	AuditCreatedAt       *time.Time             `json:"audit_created_at,omitempty" bson:"audit_created_at,omitempty"`
	AuditUpdatedAt       *time.Time             `json:"audit_updated_at,omitempty" bson:"audit_updated_at,omitempty"`
	UserID               string                 `json:"user_id,omitempty" bson:"user_id,omitempty"`
	UserType             string                 `json:"user_type,omitempty" bson:"user_type,omitempty"`
	DocumentCurrentState map[string]interface{} `json:"document_current_state,omitempty" bson:"document_current_state,omitempty"`
}

type AuditLog struct {
	Id          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AuditMetaId string             `json:"audit_meta_id,omitempty" bson:"audit_meta_id,omitempty"`
	Events      []AuditEvent       `json:"events,omitempty" bson:"events,omitempty"`
}

type AuditEvent struct {
	EventType string                 `json:"event_type,omitempty" bson:"event_type,omitempty"`
	Change    map[string]AuditChange `json:"change,omitempty" bson:"change,omitempty"`
}

type AuditChange struct {
	Old string `json:"old,omitempty" bson:"old,omitempty"`
	New string `json:"new,omitempty" bson:"new,omitempty"`
}
