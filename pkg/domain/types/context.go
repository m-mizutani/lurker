package types

import "context"

type Context struct {
	context.Context
}

func NewContext() *Context {
	ctx := &Context{
		Context: context.Background(),
	}
	return ctx
}
