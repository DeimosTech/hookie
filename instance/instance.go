package instance

// Hook interface for custom hooks
type Hook interface {
	BeforeInsert()
	AfterInsert()
}

type Inject struct {
}
