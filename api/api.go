package api

import (
	"encoding/json"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
)

type PushMsg struct {
	Cmd    comm.CMD
	Subcmd byte
	Uid    string
	Sid    string
	Rid    string
	Body   string
	Ext    *comm.PushExt
}

// Push用来给connsvr推送消息，复用clog的功能
func Push(pushMsg *PushMsg) error {
	bs, _ := json.Marshal(pushMsg)
	clog.Busi(comm.BUSI_PUSH, "%s", bs)
	return nil
}
