package in

// Hook interface for custom hooks
type Hook interface {
	BeforeInsert(model interface{})
	AfterInsert(model interface{})
}

type Inject struct {
}

type Test struct {
	*Inject
	Id   string
	Name string
}
