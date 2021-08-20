package slack

import (
	"context"
	"errors"
	"os"
	"sort"

	"github.com/slack-go/slack"
)

const (
	ColorGood    = "good"
	ColorWarning = "warning"
	ColorDanger  = "danger"
)

type Message struct {
	Channel string `json:"channel"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Link    string `json:"link"`
	// good (green), warning (yellow), danger (red), or any hex color code (eg. #439FE0)
	Color  string            `json:"color"`
	Fields map[string]string `json:"fields"`
}

type Event struct {
	MsgID   string `json:"client_msg_id"`
	Type    string `json:"type"`
	Team    string `json:"team"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

type Client interface {
	SendMessage(ctx context.Context, m Message) error
}

func NewFromEnv() (Client, error) {
	if topic := os.Getenv("SLACK_NOTIFICATIONS_TOPIC"); topic != "" {
		return NewPubSub(topic)
	}
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return nil, errors.New("SLACK_TOKEN or SLACK_NOTIFICATIONS_TOPIC must be set")
	}
	return &client{
		cli: slack.New(token),
		env: os.Getenv("SENTRY_ENVIRONMENT"),
	}, nil
}

type client struct {
	cli *slack.Client
	env string
}

func (c *client) SendMessage(ctx context.Context, m Message) error {
	if m.Channel == "" {
		return errors.New("channel name must be set")
	}
	var fields []slack.AttachmentField
	if c.env != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "environment",
			Value: os.Getenv("SENTRY_ENVIRONMENT"),
		})
	}
	var names []string
	for k := range m.Fields {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		fields = append(fields, slack.AttachmentField{
			Title: k,
			Value: m.Fields[k],
		})
	}

	_, _, _, err := c.cli.SendMessage(m.Channel, slack.MsgOptionAttachments(slack.Attachment{
		Title:      m.Title,
		TitleLink:  m.Link,
		Text:       m.Text,
		Color:      m.Color,
		Fields:     fields,
		MarkdownIn: []string{"text"},
	}))
	return err
}
