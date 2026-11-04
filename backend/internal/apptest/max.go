package apptest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"neprodlenka/internal/maxapi"
	"neprodlenka/internal/message"
)

type Edit struct {
	MessageID string
	Message   message.Message
}

type FakeMAX struct {
	mu       sync.Mutex
	members  []maxapi.Member
	sendErr  error
	err      error
	sent     []message.Message
	answered []Answer
	deleted  []string
	edits    chan Edit
}

func NewFakeMAX() *FakeMAX {
	return &FakeMAX{edits: make(chan Edit, 100)}
}

func (f *FakeMAX) Send(_ context.Context, msg message.Message) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sendErr != nil {
		return "", f.sendErr
	}
	f.sent = append(f.sent, msg)
	return fmt.Sprintf("m%d", len(f.sent)), nil
}

func (f *FakeMAX) SetSendError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sendErr = err
}

func (f *FakeMAX) SetError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

func (f *FakeMAX) Delete(_ context.Context, messageID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.deleted = append(f.deleted, messageID)
	return nil
}

func (f *FakeMAX) Deleted() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}

func (f *FakeMAX) Edit(_ context.Context, messageID string, msg message.Message) error {
	f.edits <- Edit{MessageID: messageID, Message: msg}
	return nil
}

type Answer struct {
	CallbackID   string
	Notification string
}

func (f *FakeMAX) AnswerCallback(_ context.Context, callbackID, notification string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.answered = append(f.answered, Answer{callbackID, notification})
	return nil
}

func (f *FakeMAX) Members(context.Context) ([]maxapi.Member, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.members, f.err
}

func (f *FakeMAX) SetMembers(members ...maxapi.Member) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members = members
}

func (f *FakeMAX) Sent() []message.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]message.Message(nil), f.sent...)
}

func (f *FakeMAX) Answered() []Answer {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Answer(nil), f.answered...)
}

func (f *FakeMAX) NextEdit(t *testing.T) Edit {
	t.Helper()
	select {
	case e := <-f.edits:
		return e
	case <-time.After(3 * time.Second):
		t.Fatal("no edit")
		return Edit{}
	}
}

func (f *FakeMAX) NoEdit(t *testing.T) {
	t.Helper()
	select {
	case e := <-f.edits:
		t.Errorf("unexpected edit of %s:\n%s", e.MessageID, e.Message.Text)
	case <-time.After(time.Second):
	}
}
