package firehose

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/utils"
)

type Client struct {
	svc        *firehose.Firehose
	streamName string
}

func New(region, streamName string) *Client {
	ssn := session.Must(session.NewSession())

	svc := firehose.New(ssn, aws.NewConfig().WithRegion(region))

	return &Client{
		svc:        svc,
		streamName: streamName,
	}
}

type firebaseSchema struct {
	CreatedAt time.Time `json:"created_at"`
	SrcHost   string    `json:"src_host"`
	DstHost   string    `json:"dst_host"`
	SrcPort   int       `json:"src_port"`
	DstPort   int       `json:"dst_port"`

	RecvAck bool   `json:"recv_ack"`
	Raw     []byte `json:"raw"`
	Text    string `json:"text"`
}

func (x *Client) Emit(ctx *types.Context, flow *model.TCPFlow) error {
	record := &firebaseSchema{
		CreatedAt: flow.CreatedAt,
		SrcHost:   flow.SrcHost.String(),
		DstHost:   flow.DstHost.String(),
		SrcPort:   int(flow.SrcPort),
		DstPort:   int(flow.DstPort),

		RecvAck: flow.RecvAck,
		Raw:     flow.RecvData,
		Text:    utils.ToGraph(flow.RecvData),
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return goerr.Wrap(err)
	}

	input := &firehose.PutRecordInput{
		DeliveryStreamName: &x.streamName,
		Record: &firehose.Record{
			Data: raw,
		},
	}

	if _, err := x.svc.PutRecordWithContext(ctx, input); err != nil {
		return goerr.Wrap(err)
	}

	return nil
}

func (x *Client) Close() error {
	return nil
}
