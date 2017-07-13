package encoding

import (
	"memhashd/container/store"
)

var requestMap = map[string]RequestMaker{
	"store": RequestMakerOf(store.RequestStore{}),
	"load":  RequestMakerOf(store.RequestLoad{}),
}

type RequestMaker interface {
	MakeRequest() Request
}

type RequestMakerFunc func() Request

func (fn RequestMakerFunc) MakeRequest() Request {
	return fn()
}

func RequestMakerOf(v interface{}) {
	valueType := reflect.TypeOf(v)
	return RequestMakerFunc(func() Request {
		value := reflect.New(valueType)
		return value.Interface().(RequestMaker), nil
	})
}
