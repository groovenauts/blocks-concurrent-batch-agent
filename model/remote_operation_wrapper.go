package model

type RemoteOperationWrapper interface {
	GetOriginal() interface{}
	Status() string
	Errors() *[]CloudAsyncOperationError
}
