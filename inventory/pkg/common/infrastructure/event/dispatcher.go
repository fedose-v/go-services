package event

type Event interface {
	Type() string
}

type Dispatcher interface {
	Dispatch(event Event) error
}
