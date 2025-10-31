package infrastructure

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}
