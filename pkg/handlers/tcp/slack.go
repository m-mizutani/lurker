package tcp

import (
	"fmt"
	"strings"

	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/slack-go/slack"
)

func outputSlack(ctx *types.Context, out interfaces.SlackFunc, flow *tcpFlow) {
	if len(flow.recvData) == 0 {
		return
	}

	fields := []*slack.TextBlockObject{
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Source*: %s:%d", flow.srcHost, flow.srcPort),
		},
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Target*: %s:%d", flow.dstHost, flow.dstPort),
		},
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Timestamp*: %s", flow.createdAt.Format("2006-01-02 15:04:05.000")),
		},
	}

	graph := strings.Replace(toGraph(flow.recvData), "`", "\\`", -1)
	hex := strings.Replace(toHex(flow.recvData), "`", "\\`", -1)
	msg := &slack.WebhookMessage{
		Text: "New payload captured",
		Attachments: []slack.Attachment{
			{
				Color: "##f2c744",
				Blocks: slack.Blocks{
					BlockSet: []slack.Block{
						slack.SectionBlock{
							Type:   "section",
							Fields: fields,
						},
						slack.SectionBlock{
							Type: "section",
							Text: &slack.TextBlockObject{
								Type:     slack.MarkdownType,
								Text:     fmt.Sprintf("*Hex*\n```%s```", hex),
								Verbatim: true,
							},
						},
						slack.SectionBlock{
							Type: "section",
							Text: &slack.TextBlockObject{
								Type:     slack.MarkdownType,
								Text:     fmt.Sprintf("*Payload*\n```%s```", graph),
								Verbatim: true,
							},
						},
					},
				},
			},
		},
	}

	out(ctx, msg)
}
