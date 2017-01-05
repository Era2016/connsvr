package proto

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
)

type MsgWS struct {
	MsgComm
}

func (msg *MsgWS) Encode(conn net.Conn, misc interface{}) (ok bool) {
	data, _ := json.Marshal(map[string]string{
		"cmd":    strconv.Itoa(int(msg.cmd)),
		"subcmd": strconv.Itoa(int(msg.subcmd)),
		"uid":    msg.uid,
		"sid":    msg.sid,
		"rid":    msg.rid,
		"body":   msg.body,
		"ext":    msg.ext,
	})

	c := misc.(*websocket.Conn)
	for i := 0; i < 2; i++ {
		retry := false
		func() {
			defer func() {
				if err := recover(); err != nil {
					// NOTICE
					// because of gorilla/websocket's Concurrency
					// panic "concurrent write to websocket connection"
					retry = true
				}
			}()

			err := c.WriteMessage(websocket.TextMessage, data)
			if err == nil {
				ok = true
			}
			return
		}()
		if !retry {
			break
		}
		time.Sleep(time.Nanosecond * 100)
	}

	return
}

type Rsp struct {
	http.ResponseWriter
	http.Hijacker
	Conn net.Conn
}

func (rsp *Rsp) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rsp.Conn, bufio.NewReadWriter(
		bufio.NewReader(rsp.Conn),
		bufio.NewWriter(rsp.Conn),
	), nil

}

func (msg *MsgWS) DecodeHS(br *bufio.Reader, conn net.Conn, misc interface{}) (*http.Request, bool) {
	defer func() {
		if err := recover(); err != nil {
			clog.Warn("MsgWS:DecodeHS() recover err: %v, stack: %s", err, debug.Stack())
		}
	}()

	req, err := http.ReadRequest(br)
	if err != nil {
		clog.Warn("MsgWS:ReadRequest() %v", err)
		return nil, false
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:   128,
		WriteBufferSize:  1,
		HandshakeTimeout: time.Millisecond,
		CheckOrigin: func(*http.Request) bool {
			return true
		},
		Error: func(http.ResponseWriter, *http.Request, int, error) {
		},
	}
	c, err := upgrader.Upgrade(&Rsp{Conn: conn}, req, nil)
	if err != nil {
		clog.Warn("MsgWS:Upgrade() %v", err)
		return nil, false
	}

	reflect.Indirect(reflect.ValueOf(misc)).Set(reflect.ValueOf(c))

	return req, true
}

func (msg *MsgWS) Decode(br *bufio.Reader, conn net.Conn, misc interface{}) bool {
	var values url.Values

	_misc := misc.(*interface{})
	if *_misc == nil {
		req, ok := msg.DecodeHS(br, conn, _misc)
		if !ok {
			return false
		}
		req.ParseForm()
		values = req.Form
		values.Add("Cookie", req.Header.Get("Cookie"))
	} else {
		c := (*_misc).(*websocket.Conn)
		_, message, err := c.ReadMessage()
		if err != nil {
			return false
		}

		values, err = url.ParseQuery(string(message))
		if err != nil {
			return false
		}
	}

	cmd := values.Get("cmd")
	cmd_b := comm.CMD(0)
	if cmd != "" {
		cmd_i, err := strconv.Atoi(cmd)
		if err != nil || cmd_i < 0 || cmd_i > 255 {
			clog.Error("MsgWS:Decode() err: %v, cmd: %v", err, cmd)
			return false
		}
		cmd_b = comm.CMD(cmd_i)
	}

	subcmd := values.Get("subcmd")
	subcmd_b := byte(0)
	if subcmd != "" {
		subcmd_i, err := strconv.Atoi(subcmd)
		if err != nil || subcmd_i < 0 || subcmd_i > 255 {
			clog.Error("MsgWS:Decode() err: %v, subcmd: %v", err, subcmd)
			return false
		}
		subcmd_b = byte(subcmd_i)
	}

	rid := values.Get("rid")
	uid := values.Get("uid")
	sid := values.Get("sid")
	body := values.Get("body")

	if len(rid) > 255 ||
		len(uid) > 255 ||
		len(sid) > 255 ||
		len(body) > 65535 {
		return false
	}

	msg.cmd = cmd_b
	msg.subcmd = subcmd_b
	msg.rid = rid
	msg.uid = uid
	msg.sid = sid
	msg.body = body

	cliExt := &comm.CliExt{
		Cookie: values.Get("Cookie"),
	}
	ext_bs, _ := json.Marshal(cliExt)
	msg.ext = string(ext_bs)

	return true
}
