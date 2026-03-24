package wecom

import "encoding/json"

const (
	wecomDefaultWebSocketURL = "wss://openws.work.weixin.qq.com"
	wecomCmdSubscribe        = "aibot_subscribe"
	wecomCmdPing             = "ping"
	wecomCmdMsgCallback      = "aibot_msg_callback"
	wecomCmdEventCallback    = "aibot_event_callback"
	wecomCmdRespondMsg       = "aibot_respond_msg"
	wecomCmdSendMsg          = "aibot_send_msg"
	wecomMaxContentBytes     = 20480
)

type wecomEnvelope struct {
	Cmd     string          `json:"cmd,omitempty"`
	Headers wecomHeaders    `json:"headers"`
	Body    json.RawMessage `json:"body,omitempty"`
	ErrCode int             `json:"errcode,omitempty"`
	ErrMsg  string          `json:"errmsg,omitempty"`
}

type wecomHeaders struct {
	ReqID string `json:"req_id,omitempty"`
}

type wecomCommand struct {
	Cmd     string       `json:"cmd"`
	Headers wecomHeaders `json:"headers"`
	Body    any          `json:"body,omitempty"`
}

type wecomSendMsgBody struct {
	ChatID   string                `json:"chatid"`
	ChatType uint32                `json:"chat_type,omitempty"`
	MsgType  string                `json:"msgtype"`
	Markdown *wecomMarkdownContent `json:"markdown,omitempty"`
}

type wecomRespondMsgBody struct {
	MsgType string              `json:"msgtype"`
	Stream  *wecomStreamContent `json:"stream,omitempty"`
}

type wecomStreamContent struct {
	ID      string `json:"id"`
	Finish  bool   `json:"finish"`
	Content string `json:"content,omitempty"`
}

type wecomMarkdownContent struct {
	Content string `json:"content"`
}

type wecomIncomingMessage struct {
	MsgID    string `json:"msgid"`
	AIBotID  string `json:"aibotid"`
	ChatID   string `json:"chatid,omitempty"`
	ChatType string `json:"chattype,omitempty"`
	From     struct {
		UserID string `json:"userid"`
	} `json:"from"`
	MsgType string `json:"msgtype"`
	Text    *struct {
		Content string `json:"content"`
	} `json:"text,omitempty"`
	Image *struct {
		URL    string `json:"url"`
		AESKey string `json:"aeskey,omitempty"`
	} `json:"image,omitempty"`
	File *struct {
		URL    string `json:"url"`
		AESKey string `json:"aeskey,omitempty"`
	} `json:"file,omitempty"`
	Video *struct {
		URL    string `json:"url"`
		AESKey string `json:"aeskey,omitempty"`
	} `json:"video,omitempty"`
	Voice *struct {
		Content string `json:"content"`
	} `json:"voice,omitempty"`
	Mixed *struct {
		MsgItem []struct {
			MsgType string `json:"msgtype"`
			Text    *struct {
				Content string `json:"content"`
			} `json:"text,omitempty"`
			Image *struct {
				URL    string `json:"url"`
				AESKey string `json:"aeskey,omitempty"`
			} `json:"image,omitempty"`
			File *struct {
				URL    string `json:"url"`
				AESKey string `json:"aeskey,omitempty"`
			} `json:"file,omitempty"`
		} `json:"msg_item"`
	} `json:"mixed,omitempty"`
	Quote *struct {
		MsgType string `json:"msgtype"`
		Text    *struct {
			Content string `json:"content"`
		} `json:"text,omitempty"`
	} `json:"quote,omitempty"`
	Event *struct {
		EventType string `json:"eventtype"`
	} `json:"event,omitempty"`
}

func incomingChatID(msg wecomIncomingMessage) string {
	if msg.ChatID != "" {
		return msg.ChatID
	}
	return msg.From.UserID
}

func incomingChatTypeCode(kind string) uint32 {
	if kind == "group" {
		return 2
	}
	return 1
}
