package lurker

import (
	// "errors"
	"fmt"
	"github.com/fluent/fluent-logger-golang/fluent"
	"io"
	"os"
	"time"
)

type EmitterGateway struct {
	emitters []Emitter
}

func (x *EmitterGateway) Add(emitter Emitter) {
	x.emitters = append(x.emitters, emitter)
}

func (x *EmitterGateway) Emit(tag string, msg map[string]interface{}) error {
	ts := time.Now()
	for _, emitter := range x.emitters {
		err := emitter.Emit(tag, ts, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

type Emitter interface {
	Emit(tag string, timestamp time.Time, msg map[string]interface{}) error
}

/*************************************
 Emitter for storing queue (mainly for debug)
**************************************/
type Queue struct {
	Messages []map[string]interface{}
}

// Constructor of Queue Emitter
func NewQueue() (*Queue, error) {
	return &Queue{}, nil
}

// Queue::Emit
func (x *Queue) Emit(tag string, timestamp time.Time, msg map[string]interface{}) error {
	x.Messages = append(x.Messages, msg)
	return nil
}

/*************************************
 Emitter for standard output
**************************************/
type Stdout struct {
	dst io.Writer
}

// Constructor of Stdout Emitter
func NewStdout() (*Stdout, error) {
	out := &Stdout{dst: os.Stdout}
	return out, nil
}

// Stdout::Emit
func (x *Stdout) Emit(tag string, timestamp time.Time, msg map[string]interface{}) error {
	fmt.Fprintln(x.dst, timestamp, tag, msg)
	return nil
}

/*************************************
 Emitter for fluentd
**************************************/
type Fluentd struct {
	logger *fluent.Fluent
}

// Constructor of Fluentd Emitter
func NewFluentd(host string, port int) (*Fluentd, error) {
	config := fluent.Config{FluentPort: port, FluentHost: host}
	logger, err := fluent.New(config)
	if err != nil {
		return nil, err
	}

	emitter := Fluentd{}
	emitter.logger = logger
	return &emitter, nil
}

// Fluentd::Emit
func (x *Fluentd) Emit(tag string, timestamp time.Time, msg map[string]interface{}) error {
	error := x.logger.PostWithTime(tag, timestamp, msg)
	return error
}
