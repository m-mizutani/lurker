package lurker

import (
	"errors"
)

type Emitter interface {
	Emit(tag string, msg map[string]interface{})
}

type Queue struct {
	Messages []map[string]interface{}
}

func (x *Queue) Emit(tag string, msg map[string]interface{}) {
	x.Messages = append(x.Messages, msg)
}

func NewEmiter(t string) (Emitter, error) {
	switch t {
	case "queue":
		return &Queue{}, nil
	default:
		return nil, errors.New("No such emitter: " + t)
	}
}
