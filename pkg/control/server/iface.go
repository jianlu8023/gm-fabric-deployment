package server

type Server interface {
	StartUp(failedFunc func(err error))
	Shutdown() error
}
