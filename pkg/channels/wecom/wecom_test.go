package wecom

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestDispatchIncoming_UsesActualChatIDAndStoresReqIDRoute(t *testing.T) {
	t.Parallel()

	messageBus := bus.NewMessageBus()
	ch := newTestWeComChannel(t, messageBus)

	var commands []wecomCommand
	ch.commandSend = func(cmd wecomCommand, _ time.Duration) error {
		commands = append(commands, cmd)
		return nil
	}

	msg := wecomIncomingMessage{
		MsgID:    "msg-1",
		ChatID:   "chat-1",
		ChatType: "direct",
		MsgType:  "text",
		Text: &struct {
			Content string `json:"content"`
		}{Content: "hello"},
	}
	msg.From.UserID = "user-1"

	if err := ch.dispatchIncoming("req-1", msg); err != nil {
		t.Fatalf("dispatchIncoming() error = %v", err)
	}

	select {
	case inbound := <-messageBus.InboundChan():
		if inbound.ChatID != "chat-1" {
			t.Fatalf("inbound ChatID = %q, want chat-1", inbound.ChatID)
		}
		if inbound.MessageID != "msg-1" {
			t.Fatalf("inbound MessageID = %q, want msg-1", inbound.MessageID)
		}
		if inbound.Peer.ID != "chat-1" {
			t.Fatalf("inbound Peer.ID = %q, want chat-1", inbound.Peer.ID)
		}
		if inbound.Metadata["req_id"] != "req-1" {
			t.Fatalf("inbound req_id = %q, want req-1", inbound.Metadata["req_id"])
		}
	default:
		t.Fatal("expected inbound message to be published")
	}

	turn, ok := ch.getTurn("chat-1")
	if !ok {
		t.Fatal("expected queued turn for chat-1")
	}
	if turn.ReqID != "req-1" {
		t.Fatalf("turn.ReqID = %q, want req-1", turn.ReqID)
	}

	route, ok := ch.routes.Get("chat-1")
	if !ok {
		t.Fatal("expected persisted route for chat-1")
	}
	if route.ReqID != "req-1" || route.ChatType != 1 {
		t.Fatalf("route = %+v", route)
	}

	if len(commands) != 1 {
		t.Fatalf("expected 1 opening command, got %d", len(commands))
	}
	if commands[0].Cmd != wecomCmdRespondMsg {
		t.Fatalf("opening command = %q, want %q", commands[0].Cmd, wecomCmdRespondMsg)
	}
	if commands[0].Headers.ReqID != "req-1" {
		t.Fatalf("opening req_id = %q, want req-1", commands[0].Headers.ReqID)
	}
}

func TestSend_StreamFailureFallsBackToActualChatID(t *testing.T) {
	t.Parallel()

	ch := newTestWeComChannel(t, bus.NewMessageBus())
	ch.SetRunning(true)
	ch.queueTurn("chat-1", wecomTurn{
		ReqID:     "req-1",
		ChatID:    "chat-1",
		ChatType:  1,
		StreamID:  "stream-1",
		CreatedAt: time.Now(),
	})
	ch.queueTurn("chat-1", wecomTurn{
		ReqID:     "req-2",
		ChatID:    "chat-1",
		ChatType:  1,
		StreamID:  "stream-2",
		CreatedAt: time.Now(),
	})
	if err := ch.routes.Put("chat-1", "req-2", 1, time.Hour); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	var commands []wecomCommand
	ch.commandSend = func(cmd wecomCommand, _ time.Duration) error {
		commands = append(commands, cmd)
		if len(commands) == 1 && cmd.Cmd == wecomCmdRespondMsg {
			return errors.New("stream send failed")
		}
		return nil
	}

	if err := ch.Send(context.Background(), bus.OutboundMessage{
		Channel: "wecom",
		ChatID:  "chat-1",
		Content: "hello",
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}
	if commands[0].Cmd != wecomCmdRespondMsg || commands[0].Headers.ReqID != "req-1" {
		t.Fatalf("first command = %+v", commands[0])
	}
	if commands[1].Cmd != wecomCmdSendMsg {
		t.Fatalf("second command = %q, want %q", commands[1].Cmd, wecomCmdSendMsg)
	}
	body, ok := commands[1].Body.(wecomSendMsgBody)
	if !ok {
		t.Fatalf("unexpected send body type %T", commands[1].Body)
	}
	if body.ChatID != "chat-1" {
		t.Fatalf("send chatid = %q, want chat-1", body.ChatID)
	}
	if body.ChatType != 1 {
		t.Fatalf("send chat_type = %d, want 1", body.ChatType)
	}

	nextTurn, ok := ch.getTurn("chat-1")
	if !ok {
		t.Fatal("expected second turn to remain queued")
	}
	if nextTurn.ReqID != "req-2" {
		t.Fatalf("next queued req_id = %q, want req-2", nextTurn.ReqID)
	}
}

func newTestWeComChannel(t *testing.T, messageBus *bus.MessageBus) *WeComChannel {
	t.Helper()

	cfg := config.WeComConfig{BotID: "bot-1"}
	cfg.SetSecret("secret-1")
	ch, err := NewChannel(cfg, messageBus)
	if err != nil {
		t.Fatalf("NewChannel() error = %v", err)
	}
	ch.ctx = context.Background()
	ch.routes = newReqIDStore(filepath.Join(t.TempDir(), "reqids.json"))
	return ch
}
