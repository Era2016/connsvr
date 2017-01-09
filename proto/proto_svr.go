package proto

import (
	"bufio"
	"net"
	"net/http"
	"strconv"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
)

type MsgSvr struct {
	MsgComm
}

func (msg *MsgSvr) Decode(br *bufio.Reader, conn net.Conn, misc interface{}) bool {
	req := misc.(*http.Request)

	cmd := req.FormValue("cmd")
	cmd_b := comm.CMD(0)
	if cmd != "" {
		cmd_i, err := strconv.Atoi(cmd)
		if err != nil || cmd_i < 0 || cmd_i > 255 {
			clog.Error("MsgSvr:DecodeReq() err: %v, cmd: %v", err, cmd)
			return false
		}
		cmd_b = comm.CMD(cmd_i)
	}

	subcmd := req.FormValue("subcmd")
	subcmd_b := byte(0)
	if subcmd != "" {
		subcmd_i, err := strconv.Atoi(subcmd)
		if err != nil || subcmd_i < 0 || subcmd_i > 255 {
			clog.Error("MsgSvr:DecodeReq() err: %v, subcmd: %v", err, subcmd)
			return false
		}
		subcmd_b = byte(subcmd_i)
	}

	rid := req.FormValue("rid")
	uid := req.FormValue("uid")
	sid := req.FormValue("sid")
	body := req.FormValue("body")
	ext := req.FormValue("ext")

	if len(rid) > 255 ||
		len(uid) > 255 ||
		len(sid) > 255 ||
		len(body) > 65535 ||
		len(ext) > 65535 {
		return false
	}

	msg.cmd = cmd_b
	msg.subcmd = subcmd_b
	msg.rid = rid
	msg.uid = uid
	msg.sid = sid
	msg.body = body
	msg.ext = ext

	return true
}
