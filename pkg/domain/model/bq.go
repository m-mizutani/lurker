package model

import "time"

type SchemaTcpData struct {
	ID        string
	CreatedAt time.Time
	SrcHost   string
	SrcPort   int
	DstHost   string
	DstPort   int
	ACKed     bool

	Payload string
	RawData []byte `bigquery:",nullable"`
}
