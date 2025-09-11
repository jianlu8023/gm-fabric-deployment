package server

type ControlInterface interface {
	StartUp(failedFunc func(err error))
	Shutdown() error
}
