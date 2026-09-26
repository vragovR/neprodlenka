package maxapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"

	"neprodlenka/internal/config"
	"neprodlenka/internal/message"
)

type Member struct {
	UserID int64
	Name   string
}

type Client struct {
	api    *maxbot.Api
	chatID int64
}

func New(cfg config.Config) (*Client, error) {
	api, err := maxbot.NewApi(cfg.MaxToken, maxbot.WithPollingTimeout(20*time.Second))
	if err != nil {
		return nil, fmt.Errorf("create max api: %w", err)
	}
	return &Client{api: api, chatID: cfg.MaxChatID}, nil
}

func (c *Client) Send(ctx context.Context, msg message.Message) (string, error) {
	m := maxbot.NewMessage().SetChat(c.chatID).SetText(msg.Text).SetFormat(model.FormatHTML)
	if len(msg.Buttons) > 0 {
		m.AddKeyboard(keyboard(msg.Buttons))
	}
	res, err := c.api.Messages.Send(ctx, m)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}
	return res.Message.Body.Mid, nil
}

func (c *Client) Edit(ctx context.Context, messageID string, msg message.Message) error {
	body := model.NewMessageBody{Text: msg.Text, Format: model.FormatHTML, Attachments: []model.Attachment{}}
	if len(msg.Buttons) > 0 {
		body.Attachments = append(body.Attachments, keyboard(msg.Buttons).Build())
	}
	if _, err := c.api.Messages.EditMessage(ctx, messageID, body); err != nil {
		return fmt.Errorf("edit message %s: %w", messageID, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, messageID string) error {
	res, err := c.api.Messages.DeleteMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("delete message %s: %w", messageID, err)
	}
	if !res.Success {
		return fmt.Errorf("delete message %s: %s", messageID, res.Message)
	}
	return nil
}

func (c *Client) Members(ctx context.Context) ([]Member, error) {
	var (
		members []Member
		marker  int64
	)
	for {
		list, err := c.api.Chats.GetMembers(ctx, c.chatID, marker, 100, nil)
		if err != nil {
			return nil, fmt.Errorf("get chat members: %w", err)
		}
		for _, m := range list.Members {
			if !m.IsBot {
				members = append(members, Member{
					UserID: m.UserID,
					Name:   strings.TrimSpace(m.FirstName + " " + m.LastName),
				})
			}
		}
		if list.Marker == 0 {
			return members, nil
		}
		marker = list.Marker
	}
}

func (c *Client) AnswerCallback(ctx context.Context, callbackID, notification string) error {
	if _, err := c.api.Messages.AnswerOnCallback(ctx, callbackID, model.CallbackAnswer{Notification: &notification}); err != nil {
		return fmt.Errorf("answer callback: %w", err)
	}
	return nil
}

func (c *Client) Subscribe(ctx context.Context, url, secret string) error {
	res, err := c.api.Subscriptions.Subscribe(ctx, url, secret, []string{string(model.UpdateMessageCallback)}, "")
	if err != nil {
		return fmt.Errorf("subscribe webhook: %w", err)
	}
	if !res.Success {
		return fmt.Errorf("subscribe webhook: %s", res.Message)
	}
	return nil
}

func (c *Client) WebhookHandler(secret string, handle func(context.Context, model.Update)) http.Handler {
	return c.api.GetHandler(handle, secret)
}

func (c *Client) Updates(ctx context.Context, marker int64) ([]model.Update, int64, error) {
	return c.api.Subscriptions.GetUpdates(ctx, marker)
}

func (c *Client) ChatID() int64 {
	return c.chatID
}

func keyboard(buttons []message.Button) *model.Keyboard {
	kb := model.NewKeyboard()
	row := kb.AddRow()
	for _, b := range buttons {
		row.AddCallBack(b.Text, b.Payload)
	}
	return kb
}
