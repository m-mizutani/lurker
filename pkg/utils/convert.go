package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ToGraph(data []byte) string {
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

func ToHex(data []byte) string {
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
