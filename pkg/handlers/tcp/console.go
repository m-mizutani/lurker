package tcp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
)

func outputConsole(out interfaces.ConsoleFunc, flow *tcpFlow) {
	state := "SYNed"
	if flow.recvAck {
		state = "ACKed"
	}

	hdr := fmt.Sprintf("%s:%d -> %s:%d (%s)",
		flow.srcHost.String(),
		flow.srcPort,
		flow.dstHost.String(),
		flow.dstPort,
		state,
	)

	lines := []string{hdr}

	graph := make([]rune, len(flow.recvData))
	for i := 0; i < len(flow.recvData); i += 16 {
		hex := make([]string, 16)
		asc := make([]string, 16)

		for n := 0; n < 16; n++ {
			d := i + n
			if i+n < len(flow.recvData) {
				p := flow.recvData[d]

				hex[n] = fmt.Sprintf("%02X", p)
				if strconv.IsPrint(rune(p)) {
					asc[n] = string(p)
				} else {
					asc[n] = "."
				}

				if strconv.IsGraphic(rune(p)) || rune(p) == '\n' || rune(p) == '\r' || rune(p) == '\t' {
					graph[d] = rune(p)
				} else {
					graph[d] = '.'
				}
			} else {
				hex[n] = "  "
				asc[n] = " "
			}
		}
		lines = append(lines, fmt.Sprintf(
			"%04X  %s  | %s |", i, strings.Join(hex, " "), strings.Join(asc, ""),
		))
	}

	if len(flow.recvData) > 0 {
		lines = append(lines, []string{
			"----------------------------------[dump]----------------------------------",
			string(graph),
			"--------------------------------------------------------------------------",
			"",
		}...)
	}
	out(strings.Join(lines, "\n"))
}
