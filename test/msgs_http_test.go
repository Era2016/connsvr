package tests

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestMsgsHttp(t *testing.T) {
	cmd := comm.PUSH
	subcmd := byte(1)
	rid := "r1"
	uid := "u_TestMsgsHttp"
	text := "hello world"
	msgId := "1"

	conn, err := net.Dial(
		"udp",
		fmt.Sprintf("%s:%d", utils.LocalIp, conf.C.App.Bport),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg := proto.NewMsg(comm.UDP)
	msg.SetCmd(cmd)
	msg.SetSubcmd(subcmd)
	msg.SetRid(rid)
	msg.SetUid(uid)
	msg.SetBody(text)
	pushExt := &comm.PushExt{
		MsgId:    msgId,
		PushKind: 0,
	}
	pushExt_bs, _ := json.Marshal(pushExt)
	msg.SetExt(string(pushExt_bs))
	data, ok := msg.Encode()
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	msgIds := map[byte]string{subcmd: ""}
	msgsBody := &comm.MsgsBody{
		MsgIds: msgIds,
	}
	msgsBody_bs, _ := json.Marshal(msgsBody)
	gpp := &utils.GPP{
		Uri: fmt.Sprintf("http://:%d", conf.C.App.Hport),
		Headers: map[string]string{
			"Connection": "Close",
		},
		Params: map[string]string{
			"cmd":      strconv.Itoa(int(comm.MSGS)),
			"subcmd":   "0",
			"uid":      "",
			"rid":      rid,
			"body":     string(msgsBody_bs),
			"callback": "",
		},
	}
	resp, err := utils.Get(gpp)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]string
	json.Unmarshal(resp, &m)
	if _cmd := m["cmd"]; _cmd != strconv.Itoa(int(comm.MSGS)) {
		t.Errorf("get: %v, expected: %v", _cmd, comm.MSGS)
	}

	expect_body, _ := json.Marshal(map[byte][]string{subcmd: []string{text}})
	if body := m["body"]; body != string(expect_body) {
		t.Errorf("get: %s, expected: %s", body, expect_body)
	}
}
