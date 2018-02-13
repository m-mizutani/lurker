package lurker

import (
	"errors"
	"time"
)

type EmitterGateway struct {
	emitters []Emitter
}

func (x *EmitterGateway) Add(emitter Emitter) {
	x.emitters = append(x.emitters, emitter)
}

func (x *EmitterGateway) Emit(tag string, msg map[string]interface{}) {
	ts := int(time.Now().Unix())
	for _, emitter := range x.emitters {
		emitter.Emit(tag, ts, msg)
	}
}

type Emitter interface {
	Emit(tag string, timestamp int, msg map[string]interface{})
}

type Queue struct {
	Messages []map[string]interface{}
}

func (x *Queue) Emit(tag string, timestamp int, msg map[string]interface{}) {
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
