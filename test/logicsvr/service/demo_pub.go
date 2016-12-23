package service

import (
	"github.com/simplejia/connsvr/api"
	"github.com/simplejia/connsvr/comm"
)

func (demo *Demo) Pub(rid, body string) (err error) {
	msg := &api.PushMsg{
		Cmd:    comm.PUSH,
		Subcmd: 1,
		Uid:    "",
		Sid:    "",
		Rid:    rid,
		Body:   body,
		Ext: &comm.PushExt{
			MsgId:    "1",
			PushKind: comm.DISPLAY,
		},
	}
	api.Push(msg)
	return
}
