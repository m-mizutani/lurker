package slack

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/utils"
	"github.com/slack-go/slack"
)

type Emitter struct {
	webhookURL string
}

func New(webhookURL string) *Emitter {
	return &Emitter{
		webhookURL: webhookURL,
	}
}

func (x *Emitter) Emit(ctx *types.Context, flow *model.TCPFlow) error {
	if len(flow.RecvData) == 0 {
		return nil
	}

	fields := []*slack.TextBlockObject{
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Source*: %s:%d", flow.SrcHost, flow.SrcPort),
		},
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Target*: %s:%d", flow.DstHost, flow.DstPort),
		},
		{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("*Timestamp*: %s", flow.CreatedAt.Format("2006-01-02 15:04:05.000")),
		},
	}

	graph := strings.Replace(utils.ToGraph(flow.RecvData), "`", "\\`", -1)
	hex := strings.Replace(utils.ToHex(flow.RecvData), "`", "\\`", -1)
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

	if err := slack.PostWebhookContext(ctx, x.webhookURL, msg); err != nil {
		raw, _ := json.Marshal(msg)
		return goerr.Wrap(err).With("body", string(raw))
	}

	return nil
}

func (x *Emitter) Close() error {
	return nil
}
