package console

import (
	"fmt"
	"io"
	"strings"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/utils"
)

type Emitter struct {
	w io.Writer
}

func New(w io.Writer) *Emitter {
	return &Emitter{
		w: w,
	}
}

func (x *Emitter) Emit(ctx *types.Context, flow *model.TCPFlow) error {
	state := "SYNed"
	if flow.RecvAck {
		state = "ACKed"
	}

	hdr := fmt.Sprintf("%s:%d -> %s:%d (%s)",
		flow.SrcHost.String(),
		flow.SrcPort,
		flow.DstHost.String(),
		flow.DstPort,
		state,
	)

	lines := []string{hdr, utils.ToHex(flow.RecvData)}

	if len(flow.RecvData) > 0 {
		graph := utils.ToGraph(flow.RecvData)
		lines = append(lines, []string{
			"----------------------------------[dump]----------------------------------",
			string(graph),
			"--------------------------------------------------------------------------",
			"",
		}...)
	}

	if _, err := x.w.Write([]byte(strings.Join(lines, "\n"))); err != nil {
		return goerr.Wrap(err)
	}

	return nil
}

func Close() error {
	// Do not close writer to prevent unexpected closing output. e.g. stdout
	return nil
}
