package bq

import (
	"context"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
)

const (
	tcpDataTable = "tcp_data"
)

type Client struct {
	projectID string
	dataSet   string

	client  *bigquery.Client
	tcpData *bigquery.Table
}

func New(projectID, dataSet string) (*Client, error) {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	ds := client.Dataset(dataSet)

	tcpData := ds.Table(tcpDataTable)
	schema, err := bigquery.InferSchema(model.SchemaTcpData{})
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

	return &Client{
		projectID: projectID,
		dataSet:   dataSet,
		client:    client,
		tcpData:   tcpData,
	}, nil
}

func (x *Client) InsertTcpData(ctx *types.Context, data *model.SchemaTcpData) error {
	inserter := x.tcpData.Inserter()
	if err := inserter.Put(ctx, []*model.SchemaTcpData{data}); err != nil {
		return goerr.Wrap(err)
	}
	return nil
}
