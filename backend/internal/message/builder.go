package message

import (
	"context"
	"slices"
	"time"

	"neprodlenka/internal/config"
	"neprodlenka/internal/report"
	"neprodlenka/internal/store"
)

type Builder struct {
	store     *store.Store
	closeTime string
	payment   Payment
	baseURL   string
}

func NewBuilder(s *store.Store, cfg config.Config) *Builder {
	return &Builder{store: s, closeTime: cfg.CloseTime.String(), payment: Payment{Card: cfg.PaymentCard, Recipient: cfg.PaymentRecipient}, baseURL: cfg.BaseURL}
}

func (b *Builder) Build(ctx context.Context, date time.Time) (Message, store.Post, bool, error) {
	post, ok, err := b.store.Post(ctx, date)
	if err != nil || !ok {
		return Message{}, store.Post{}, false, err
	}
	children, err := b.dayChildren(ctx, date, true)
	if err != nil {
		return Message{}, store.Post{}, false, err
	}
	clubs, err := b.store.DayClubs(ctx, date)
	if err != nil {
		return Message{}, store.Post{}, false, err
	}
	msg, err := b.buildDay(ctx, date, post.Status == store.PostClosed, children, clubs)
	return msg, post, err == nil, err
}

func (b *Builder) BuildDay(ctx context.Context, date time.Time, closed, marked bool) (Message, error) {
	children, err := b.dayChildren(ctx, date, marked)
	if err != nil {
		return Message{}, err
	}
	clubs, err := b.store.PlannedClubs(ctx, date)
	if err != nil {
		return Message{}, err
	}
	if marked {
		clubs = slices.DeleteFunc(clubs, func(club store.DayClub) bool {
			return !slices.ContainsFunc(children, func(c store.DayChild) bool { return c.ID == club.ChildID })
		})
	}
	return b.buildDay(ctx, date, closed, children, clubs)
}

func (b *Builder) dayChildren(ctx context.Context, date time.Time, marked bool) ([]store.DayChild, error) {
	children, err := b.store.DayChildren(ctx, date)
	if err != nil || !marked {
		return children, err
	}
	return slices.DeleteFunc(children, func(c store.DayChild) bool { return c.Status == "" }), nil
}

func (b *Builder) buildDay(ctx context.Context, date time.Time, closed bool, children []store.DayChild, clubs []store.DayClub) (Message, error) {
	uid, err := b.store.ReportUID(ctx, date)
	if err != nil {
		return Message{}, err
	}
	return Render(Day{
		Date:      date,
		CloseTime: b.closeTime,
		Closed:    closed,
		Children:  children,
		Clubs:     clubs,
		URL:       b.reportURL(uid),
	}), nil
}

func (b *Builder) BuildMonth(ctx context.Context, month time.Time) (Message, store.ReportMonth, bool, error) {
	rm, err := b.store.ReportMonth(ctx, month)
	if err != nil {
		return Message{}, store.ReportMonth{}, false, err
	}
	marks, err := b.store.MonthMarks(ctx, month)
	if err != nil || len(marks) == 0 {
		return Message{}, store.ReportMonth{}, false, err
	}
	days, err := b.store.NoSchoolDays(ctx, month)
	if err != nil {
		return Message{}, store.ReportMonth{}, false, err
	}
	children, err := b.store.Children(ctx)
	if err != nil {
		return Message{}, store.ReportMonth{}, false, err
	}
	paid, err := b.store.MonthPaid(ctx, month)
	if err != nil {
		return Message{}, store.ReportMonth{}, false, err
	}
	r := report.Build(month, days, children, marks)
	return RenderMonth(month, r, paid, b.reportURL(rm.UID), b.payment), rm, true, nil
}

func (b *Builder) reportURL(uid string) string {
	return b.baseURL + "/report/" + uid
}
