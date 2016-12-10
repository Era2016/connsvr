package tests

import (
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestHttpPub(t *testing.T) {
	rid := "r1"
	uid := "u_TestHttpPub"
	sid := "s1"
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
	msg.SetCmd(comm.ENTER)
	msg.SetUid(uid)
	msg.SetRid(rid)
	data, ok := msg.Encode()
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	gpp := &utils.GPP{
		Uri: fmt.Sprintf("http://:%d", conf.C.App.Hport),
		Headers: map[string]string{
			"Connection": "Close",
		},
		Params: map[string]string{
			"cmd":      strconv.Itoa(int(comm.PUB)),
			"subcmd":   "1",
			"rid":      rid,
			"uid":      uid,
			"sid":      sid,
			"body":     text,
			"callback": "",
		},
	}
	resp, err := utils.Get(gpp)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]string
	json.Unmarshal(resp, &m)
	if _cmd := m["cmd"]; _cmd == strconv.Itoa(int(comm.ERR)) {
		t.Errorf("get: %v, expected: %v", _cmd, comm.PUB)
		t.Errorf("please check you conf(pubs)!!!")
	}

	t.Log("get resp:", m["body"])
}
