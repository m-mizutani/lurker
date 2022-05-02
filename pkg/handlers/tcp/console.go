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

	lines := []string{hdr, toHex(flow.recvData)}

	if len(flow.recvData) > 0 {
		graph := toGraph(flow.recvData)
		lines = append(lines, []string{
			"----------------------------------[dump]----------------------------------",
			string(graph),
			"--------------------------------------------------------------------------",
			"",
		}...)
	}
	out(strings.Join(lines, "\n"))
}

func toHex(data []byte) string {
	lines := []string{}

	for i := 0; i < len(data); i += 16 {
		hex := make([]string, 16)
		asc := make([]string, 16)

		for n := 0; n < 16; n++ {
			d := i + n
			if i+n < len(data) {
				p := data[d]

				hex[n] = fmt.Sprintf("%02X", p)
				if strconv.IsPrint(rune(p)) {
					asc[n] = string(p)
				} else {
					asc[n] = "."
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

	return strings.Join(lines, "\n")
}

func toGraph(data []byte) string {
	graph := make([]rune, len(data))
	for d := 0; d < len(data); d++ {
		p := data[d]

		if strconv.IsGraphic(rune(p)) || rune(p) == '\n' || rune(p) == '\r' || rune(p) == '\t' {
			graph[d] = rune(p)
		} else {
			graph[d] = '.'
		}
	}

	return string(graph)
}
