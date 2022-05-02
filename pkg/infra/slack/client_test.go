package slack_test

import (
	"os"
	"testing"

	"github.com/m-mizutani/lurker/pkg/domain/types"
	s "github.com/m-mizutani/lurker/pkg/infra/slack"

	"github.com/slack-go/slack"
)

func TestSlack(t *testing.T) {
	url, ok := os.LookupEnv("LURKER_SLACK_WEBHOOK")
	if !ok {
		t.Skip("LURKER_SLACK_WEBHOOK is not set")
	}

	client := s.New(url)
	client.Post(types.NewContext(), &slack.WebhookMessage{
		Text: "test",
		Attachments: []slack.Attachment{
			{
				Color: "##f2c744",
				Blocks: slack.Blocks{
					BlockSet: []slack.Block{
						slack.SectionBlock{
							Type: "section",
							Text: &slack.TextBlockObject{
								Type: slack.MarkdownType,
								Text: "hello",
							},
						},
					},
				},
			},
		},
	})
}
