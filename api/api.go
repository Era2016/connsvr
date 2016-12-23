package api

import (
	"encoding/json"
	"net"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
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

// Push用来给connsvr推送消息，复用clog的功能
func PushRaw(pushMsg *PushMsg, ips []string) error {
	msg := proto.NewMsg(comm.UDP)
	msg.SetCmd(pushMsg.Cmd)
	msg.SetSubcmd(pushMsg.Subcmd)
	msg.SetUid(pushMsg.Uid)
	msg.SetSid(pushMsg.Sid)
	msg.SetRid(pushMsg.Rid)
	msg.SetBody(pushMsg.Body)
	if pushMsg.Ext != nil {
		bs, _ := json.Marshal(pushMsg.Ext)
		msg.SetExt(string(bs))
	}
	data, ok := msg.Encode()
	if !ok {
		clog.Error("ConnPushHandler() msg encode error, msg: %+v", msg)
		return nil
	}

	for _, ipport := range ips {
		conn, err := net.Dial("udp", ipport)
		if err != nil {
			clog.Error("ConnPushHandler() dial ipport: %s, error: %v", ipport, err)
			continue
		}
		defer conn.Close()

		_, err = conn.Write(data)
		if err != nil {
			clog.Error("ConnPushHandler() conn.Write ipport: %s, error: %v", ipport, err)
			continue
		}
	}

	return nil
}
