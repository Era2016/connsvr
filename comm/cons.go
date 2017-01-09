//go:generate stringer -type=CMD,PUSH_KIND,PROTO -output=cons_string.go

package comm

import "time"

type CMD byte

const (
	PING CMD = iota + 1 // 1, 现在的技术方案用不到心跳
	ENTER
	LEAVE
	PUB
	MSGS
	PUSH
	ERR = 0xff
)

type PUSH_KIND byte

const (
	NOTIFY  PUSH_KIND = iota + 1 // 推送通知，然后客户端主动拉connsvr或者后端服务
	DISPLAY                      // 推送整条消息，客户端不用拉
)

type PROTO int

const (
	TCP PROTO = iota + 1 //1
	HTTP
	SVR
	WS
)

const (
	BUSI_REPORT = "report"
	BUSI_STAT   = "stat"
	BUSI_PUSH   = "push"
)

type Stat struct {
	Ip    string
	N     int
	Rid   string
	Msg   string
	Num   int
	Btime time.Time
	Etime time.Time
}

// Msgs is from logic svr
type Msgs []*struct {
	MsgId string
	Uid   string
	Sid   string
	Body  string
}

// ServExt will be transfered to client
type ServExt struct {
	PushKind PUSH_KIND
}

// PushExt is from backend
type PushExt struct {
	MsgId    string
	PushKind PUSH_KIND
}

// CliExt is from client
type CliExt struct {
	Cookie string
}

// EnterBody is from ENTER CMD
type EnterBody struct {
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}

// MsgsBody is from MSGS CMD
type MsgsBody struct {
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}
