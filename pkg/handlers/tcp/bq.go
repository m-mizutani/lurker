package tcp

import (
	"github.com/google/uuid"
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
)

func outputBigQuery(ctx *types.Context, out interfaces.InsertTcpDataFunc, flow *tcpFlow) {
	data := &model.SchemaTcpData{
		ID:        uuid.NewString(),
		CreatedAt: flow.createdAt,
		SrcHost:   flow.srcHost.String(),
		SrcPort:   int(flow.srcPort),
		DstHost:   flow.dstHost.String(),
		DstPort:   int(flow.dstPort),
		ACKed:     flow.recvAck,

		Payload: toGraph(flow.recvData),
		RawData: flow.recvData,
	}

	out(ctx, data)
}
