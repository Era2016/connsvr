package tests

import (
	"bufio"
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

func TestMsgsTcp(t *testing.T) {
	subcmd := byte(1)
	rid := "r1"
	uid := "u_TestMsgsTcp"
	text := "hello world"
	msgId := "1"

	pushExt := &comm.PushExt{
		MsgId:    msgId,
		PushKind: 0,
	}
	ext_bs, _ := json.Marshal(pushExt)

	gpp := &utils.GPP{
		Uri: fmt.Sprintf("http://%s:%d", utils.LocalIp, conf.C.App.Bport),
		Params: map[string]string{
			"cmd":    strconv.Itoa(int(comm.PUSH)),
			"subcmd": strconv.Itoa(int(subcmd)),
			"rid":    rid,
			"uid":    uid,
			"body":   text,
			"ext":    string(ext_bs),
		},
	}
	_, err := utils.Post(gpp)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	conn, err := net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", utils.LocalIp, conf.C.App.Tport),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg := proto.NewMsg(comm.TCP)
	msg.SetCmd(comm.MSGS)
	msg.SetSubcmd(0)
	msg.SetUid("")
	msg.SetRid(rid)
	msgIds := map[byte]string{subcmd: ""}
	msgsBody := &comm.MsgsBody{
		MsgIds: msgIds,
	}
	msgsBody_bs, _ := json.Marshal(msgsBody)
	msg.SetBody(string(msgsBody_bs))
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
		t.Errorf("get: %v, expected: %v", _msg.Cmd(), msg.Cmd())
	}
	if _msg.Rid() != rid {
		t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
	}

	expect_body, _ := json.Marshal(map[byte][]string{subcmd: []string{text}})
	if body := _msg.Body(); body != string(expect_body) {
		t.Errorf("get: %s, expected: %s", _msg.Body(), expect_body)
	}
}
