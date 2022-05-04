package bq_test

import (
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/infra/bq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigQueryClient(t *testing.T) {
	projectID, ok := os.LookupEnv("LURKER_BIGQUERY_PROJECT_ID")
	if !ok {
		t.Skip("LURKER_BIGQUERY_PROJECT_ID is not set")
	}
	dataset, ok := os.LookupEnv("LURKER_BIGQUERY_DATASET")
	if !ok {
		t.Skip("LURKER_BIGQUERY_DATASET is not set")
	}

	client, err := bq.New(projectID, dataset)
	require.NoError(t, err)
	assert.NotNil(t, client)

	ctx := types.NewContext()
	require.NoError(t, client.InsertTcpData(ctx, &model.SchemaTcpData{
		CreatedAt: time.Now(),
		SrcHost:   "192.168.0.1",
		SrcPort:   12345,
		DstHost:   "10.0.0.1",
		DstPort:   4321,
		Payload:   "GET",
		RawData:   []byte("GET"),
	}))

	require.NoError(t, client.InsertTcpData(ctx, &model.SchemaTcpData{
		CreatedAt: time.Now(),
		SrcHost:   "192.168.0.2",
		SrcPort:   12345,
		DstHost:   "10.0.0.2",
		DstPort:   4321,
	}))
}
