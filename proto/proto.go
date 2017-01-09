package proto

import (
	"bufio"
	"fmt"
	"net"

	"github.com/simplejia/connsvr/comm"
)

type Msg interface {
	Length() int
	SetLength(int)
	Cmd() comm.CMD
	SetCmd(comm.CMD)
	Subcmd() byte
	SetSubcmd(byte)
	Uid() string
	SetUid(string)
	Sid() string
	SetSid(string)
	Rid() string
	SetRid(string)
	Body() string
	SetBody(string)
	Ext() string
	SetExt(string)
	Encode(net.Conn, interface{}) bool
	Decode(*bufio.Reader, net.Conn, interface{}) bool
}

type MsgComm struct {
	length int
	cmd    comm.CMD
	subcmd byte
	uid    string
	sid    string
	rid    string
	body   string
	ext    string
}

func (msg *MsgComm) SetLength(length int) {
	msg.length = length
}

func (msg *MsgComm) Length() int {
	return msg.length
}

func (msg *MsgComm) Cmd() comm.CMD {
	return msg.cmd
}

func (msg *MsgComm) SetCmd(cmd comm.CMD) {
	msg.cmd = cmd
}

func (msg *MsgComm) Subcmd() byte {
	return msg.subcmd
}

func (msg *MsgComm) SetSubcmd(subcmd byte) {
	msg.subcmd = subcmd
}

func (msg *MsgComm) Body() string {
	return msg.body
}

func (msg *MsgComm) SetBody(body string) {
	msg.body = body
}

func (msg *MsgComm) Ext() string {
	return msg.ext
}

func (msg *MsgComm) SetExt(ext string) {
	msg.ext = ext
}

func (msg *MsgComm) Uid() string {
	return msg.uid
}

func (msg *MsgComm) SetUid(uid string) {
	msg.uid = uid
}

func (msg *MsgComm) Sid() string {
	return msg.sid
}

func (msg *MsgComm) SetSid(sid string) {
	msg.sid = sid
}

func (msg *MsgComm) Rid() string {
	return msg.rid
}

func (msg *MsgComm) SetRid(rid string) {
	msg.rid = rid
}

func (msg *MsgComm) Encode(net.Conn, interface{}) bool {
	return false
}

func (msg *MsgComm) Decode(*bufio.Reader, net.Conn, interface{}) bool {
	return false
}

func NewMsg(t comm.PROTO) Msg {
	switch t {
	case comm.TCP:
		return new(MsgTcp)
	case comm.HTTP:
		return new(MsgHttp)
	case comm.SVR:
		return new(MsgSvr)
	case comm.WS:
		return new(MsgWS)
	default:
		panic(fmt.Sprintf("NewMsg() not support proto: %v", t))
	}
}
