package tests

import (
	"bufio"
	"fmt"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestTcpPub(t *testing.T) {
	rid := "r1"
	uid := "u_TestTcpPub"
	text := "hello world"

	conn, err := net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", utils.LocalIp, conf.C.App.Tport),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg := proto.NewMsg(comm.TCP)
	msg.SetCmd(comm.PUB)
	msg.SetSubcmd(1)
	msg.SetUid(uid)
	msg.SetRid(rid)
	msg.SetBody(text)
	ok := msg.Encode(conn, nil)
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_msg := new(proto.MsgTcp)
	ok = _msg.Decode(bufio.NewReader(conn), nil, nil)
	if !ok {
		t.Fatal("_msg.DecodeHeader() error")
	}

	if _msg.Cmd() == comm.ERR {
		t.Logf("get: %v, expected: %v", _msg.Cmd(), msg.Cmd())
		t.Logf("please check you conf(pubs)!!!")
	}

	t.Log("get resp:", _msg.Body())
}
