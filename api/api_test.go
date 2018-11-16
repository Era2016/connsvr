package api

import (
	"testing"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/connsvr/comm"
)

func TestPush(t *testing.T) {
	clog.Init("logicsvr", "", 14, 2)

	msg := &PushMsg{
		Cmd:    comm.PUSH,
		Subcmd: 1,
		Uid:    "",
		Sid:    "",
		Rid:    "r1",
		Body:   "Hello World!",
		Ext: &comm.PushExt{
			MsgId:    "1",
			PushKind: comm.DISPLAY,
		},
	}
	Push(msg)
}
