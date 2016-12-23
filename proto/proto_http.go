package proto

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"

	"fmt"
)

type MsgHttp struct {
	MsgComm
}

func (msg *MsgHttp) Encode() ([]byte, bool) {
	data, _ := json.Marshal(map[string]string{
		"cmd":    strconv.Itoa(int(msg.cmd)),
		"subcmd": strconv.Itoa(int(msg.subcmd)),
		"uid":    msg.uid,
		"sid":    msg.sid,
		"rid":    msg.rid,
		"body":   msg.body,
		"ext":    msg.ext,
	})
	var resp []byte
	if callback, ok := msg.misc.(string); ok && callback != "" {
		resp = append(resp, callback...)
		resp = append(resp, '(')
		resp = append(resp, data...)
		resp = append(resp, ')')
	} else {
		resp = data
	}
	return []byte(
		fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: application/json;charset=UTF-8\r\n"+
			"Connection: Keep-Alive\r\n"+
			"Content-Length: %d\r\n\r\n%s",
			len(resp), resp,
		)), true
}

func (msg *MsgHttp) Decode(buf *bufio.Reader) (ok bool) {
	req, err := http.ReadRequest(buf)
	if err != nil {
		clog.Warn("MsgHttp:ReadRequest() %v", err)
		return false
	}

	cmd := req.FormValue("cmd")
	cmd_b := comm.CMD(0)
	if cmd != "" {
		cmd_i, err := strconv.Atoi(cmd)
		if err != nil || cmd_i < 0 || cmd_i > 255 {
			clog.Error("MsgHttp:DecodeReq() err: %v, cmd: %v", err, cmd)
			return false
		}
		cmd_b = comm.CMD(cmd_i)
	}

	subcmd := req.FormValue("subcmd")
	subcmd_b := byte(0)
	if subcmd != "" {
		subcmd_i, err := strconv.Atoi(subcmd)
		if err != nil || subcmd_i < 0 || subcmd_i > 255 {
			clog.Error("MsgHttp:DecodeReq() err: %v, subcmd: %v", err, subcmd)
			return false
		}
		subcmd_b = byte(subcmd_i)
	}

	rid := req.FormValue("rid")
	uid := req.FormValue("uid")
	sid := req.FormValue("sid")
	callback := req.FormValue("callback")
	body := req.FormValue("body")

	if len(rid) > 255 ||
		len(uid) > 255 ||
		len(sid) > 255 ||
		len(body) > 65535 ||
		len(callback) > 255 {
		return false
	}

	msg.cmd = cmd_b
	msg.subcmd = subcmd_b
	msg.rid = rid
	msg.uid = uid
	msg.sid = sid
	msg.body = body
	msg.misc = callback

	cliExt := &comm.CliExt{
		Cookie: req.Header.Get("Cookie"),
	}
	ext_bs, _ := json.Marshal(cliExt)
	msg.ext = string(ext_bs)
	return true
}
