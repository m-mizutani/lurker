package model

import "fmt"

type LogFunc func(format string, args ...any) error

type Output struct {
	Log LogFunc
}

func NewOutput(options ...OutputOption) *Output {
	output := &Output{
		Log: func(format string, args ...any) error {
			fmt.Printf(format, args...)
			return nil
		},
	}

	for _, opt := range options {
		opt(output)
	}

	return output
}

type OutputOption func(*Output)

func WithLog(logger LogFunc) OutputOption {
	return func(o *Output) {
		o.Log = logger
	}
}
