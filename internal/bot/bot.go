package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"go.uber.org/zap"

	"neprodlenka/internal/clock"
	"neprodlenka/internal/config"
	"neprodlenka/internal/maxapi"
	"neprodlenka/internal/message"
	"neprodlenka/internal/store"
)

type MAX interface {
	Send(ctx context.Context, msg message.Message) (string, error)
	Edit(ctx context.Context, messageID string, msg message.Message) error
	Delete(ctx context.Context, messageID string) error
	AnswerCallback(ctx context.Context, callbackID, notification string) error
	Members(ctx context.Context) ([]maxapi.Member, error)
}

var ErrSend = errors.New("send to MAX")

type Callback struct {
	ID        string
	MessageID string
	UserID    int64
	Payload   string
}

type Bot struct {
	store        *store.Store
	builder      *message.Builder
	max          MAX
	redrawer     *Redrawer
	now          clock.Clock
	chatID       int64
	postTime     config.TimeOfDay
	publishUntil config.TimeOfDay
	closeTime    config.TimeOfDay
	closeUntil   config.TimeOfDay
	summaryTime  config.TimeOfDay
	summaryUntil config.TimeOfDay
	log          *zap.Logger
}

func New(s *store.Store, b *message.Builder, m MAX, r *Redrawer, now clock.Clock, cfg config.Config, log *zap.Logger) *Bot {
	return &Bot{
		store:        s,
		builder:      b,
		max:          m,
		redrawer:     r,
		now:          now,
		chatID:       cfg.MaxChatID,
		postTime:     cfg.PostTime,
		publishUntil: cfg.PublishUntil,
		closeTime:    cfg.CloseTime,
		closeUntil:   cfg.CloseUntil,
		summaryTime:  cfg.SummaryTime,
		summaryUntil: cfg.SummaryUntil,
		log:          log,
	}
}

func (b *Bot) Tick(ctx context.Context) error {
	now := b.now()
	date := clock.Today(now)

	if !now.Before(b.closeTime.On(date)) && now.Before(b.closeUntil.On(date)) {
		post, ok, err := b.store.Post(ctx, date)
		if err != nil {
			return err
		}
		if ok && post.Status == store.PostOpen && post.MessageID != "" {
			if err := b.close(ctx, post); err != nil {
				return err
			}
		}
	}

	if !now.Before(b.summaryTime.On(date)) && now.Before(b.summaryUntil.On(date)) {
		last, err := b.store.IsLastSchoolDay(ctx, date)
		if err != nil {
			return err
		}
		if last {
			if _, err := b.PublishSummary(ctx, clock.MonthOf(date)); err != nil {
				return err
			}
		}
	}

	if now.Before(b.postTime.On(date)) || !now.Before(b.publishUntil.On(date)) {
		return nil
	}
	school, err := b.store.IsSchoolDay(ctx, date)
	if err != nil || !school {
		return err
	}
	_, err = b.Publish(ctx, date)
	return err
}

func (b *Bot) Publish(ctx context.Context, date time.Time) (bool, error) {
	created, err := b.store.CreatePost(ctx, date)
	if err != nil || !created {
		return false, err
	}

	messageID, err := b.send(ctx, date)
	if err != nil {
		return false, errors.Join(err, b.store.DeletePost(ctx, date))
	}
	if err := b.store.SetPostMessageID(ctx, date, messageID); err != nil {
		return false, err
	}

	b.log.Info("post published", zap.Time("date", date), zap.String("message_id", messageID))
	return true, nil
}

func (b *Bot) PublishSummary(ctx context.Context, month time.Time) (bool, error) {
	msg, rm, ok, err := b.builder.BuildMonth(ctx, month)
	if err != nil || !ok || rm.MessageID != "" {
		return false, err
	}
	messageID, err := b.max.Send(ctx, msg)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrSend, err)
	}
	if err := b.store.SetReportMessageID(ctx, month, messageID); err != nil {
		return false, err
	}
	b.log.Info("summary published", zap.Time("month", month), zap.String("message_id", messageID))
	return true, nil
}

func (b *Bot) send(ctx context.Context, date time.Time) (string, error) {
	msg, _, _, err := b.builder.Build(ctx, date)
	if err != nil {
		return "", err
	}
	id, err := b.max.Send(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSend, err)
	}
	return id, nil
}

func (b *Bot) close(ctx context.Context, p store.Post) error {
	if err := b.store.ClosePost(ctx, p.Date); err != nil {
		return err
	}
	b.log.Info("post closed", zap.Time("date", p.Date))
	b.redrawer.Redraw(p.Date)
	return nil
}

func (b *Bot) HandleUpdate(ctx context.Context, u model.Update) {
	fields := []zap.Field{zap.String("type", string(u.UpdateType)), zap.Int64("chat_id", u.ChatID)}
	cb := u.GetCallback()
	if u.UpdateType == model.UpdateMessageCallback {
		fields = append(fields, zap.String("payload", cb.Payload), zap.Int64("user_id", cb.User.UserID), zap.String("message_id", u.MessageID))
	}
	b.log.Info("update", fields...)
	if u.UpdateType != model.UpdateMessageCallback || u.ChatID != b.chatID {
		return
	}
	err := b.HandleCallback(ctx, Callback{
		ID:        cb.CallbackID,
		MessageID: u.MessageID,
		UserID:    cb.User.UserID,
		Payload:   cb.Payload,
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		b.log.Error("handle callback", zap.Error(err))
	}
}

func (b *Bot) HandleCallback(ctx context.Context, cb Callback) error {
	notification, err := b.toggle(ctx, cb)
	if err := b.max.AnswerCallback(ctx, cb.ID, notification); err != nil {
		b.log.Warn("answer callback", zap.Error(err))
	}
	return err
}

func (b *Bot) toggle(ctx context.Context, cb Callback) (string, error) {
	if store.Status(cb.Payload) != store.Absent {
		b.log.Info("unknown payload", zap.String("payload", cb.Payload))
		return "Кнопка устарела", nil
	}

	date, absent, ok, err := b.store.ToggleAbsent(ctx, cb.MessageID, cb.UserID)
	if err != nil {
		return "Не удалось сохранить отметку, попробуйте ещё раз", fmt.Errorf("handle callback: %w", err)
	}
	if !ok {
		b.log.Info("callback ignored", zap.Int64("user_id", cb.UserID), zap.String("message_id", cb.MessageID))
		return "Отметка не принята: день уже закрыт или вы не привязаны к ребёнку", nil
	}

	b.log.Info("absence toggled", zap.Int64("user_id", cb.UserID), zap.Time("date", date), zap.Bool("absent", absent))
	b.redrawer.Redraw(date)
	if absent {
		return "Отмечено: не обедает", nil
	}
	return "Отметка снята: обедает", nil
}

func (b *Bot) nextTick(now time.Time) time.Time {
	date := clock.Today(now)
	times := []time.Time{b.postTime.On(date), b.closeTime.On(date), b.summaryTime.On(date)}
	next := b.postTime.On(date.AddDate(0, 0, 1))
	for _, t := range times {
		if t.After(now) && t.Before(next) {
			next = t
		}
	}
	return next
}
