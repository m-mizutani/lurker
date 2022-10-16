package bq

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/utils"
	"google.golang.org/api/googleapi"
)

type schemaTcpData struct {
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

type Emitter struct {
	projectID string
	dataSet   string

	client  *bigquery.Client
	tcpData *bigquery.Table
}

const (
	tcpDataTable = "tcp_data"
)

func New(projectID, dataSet string) (*Emitter, error) {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	ds := client.Dataset(dataSet)

	tcpData := ds.Table(tcpDataTable)
	schema, err := bigquery.InferSchema(schemaTcpData{})
	if err != nil {
		return nil, err
	}

	meta := &bigquery.TableMetadata{
		Schema: schema,
		TimePartitioning: &bigquery.TimePartitioning{
			Type:  bigquery.DayPartitioningType,
			Field: "CreatedAt",
		},
	}
	if err := tcpData.Create(ctx, meta); err != nil {
		if gerr, ok := err.(*googleapi.Error); !ok || gerr.Code != 409 {
			return nil, err
		}
	}

	return &Emitter{
		projectID: projectID,
		dataSet:   dataSet,
		client:    client,
		tcpData:   tcpData,
	}, nil
}

func (x *Emitter) Emit(ctx *types.Context, flow *model.TCPFlow) error {
	inserter := x.tcpData.Inserter()

	data := &schemaTcpData{
		ID:        uuid.NewString(),
		CreatedAt: flow.CreatedAt,
		SrcHost:   flow.SrcHost.String(),
		SrcPort:   int(flow.SrcPort),
		DstHost:   flow.DstHost.String(),
		DstPort:   int(flow.DstPort),
		ACKed:     flow.RecvAck,

		Payload: utils.ToGraph(flow.RecvData),
		RawData: flow.RecvData,
	}
	if err := inserter.Put(ctx, []*schemaTcpData{data}); err != nil {
		return goerr.Wrap(err)
	}

	return nil
}

func (x *Emitter) Close() error {
	if err := x.client.Close(); err != nil {
		return goerr.Wrap(err, "failed to close BigQuery client")
	}
	return nil
}
