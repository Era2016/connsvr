package tests

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/utils"

	"testing"
)

func TestMsgsHttp(t *testing.T) {
	subcmd := byte(1)
	rid := "r1"
	uid := "u_TestMsgsHttp"
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

	msgIds := map[byte]string{subcmd: ""}
	msgsBody := &comm.MsgsBody{
		MsgIds: msgIds,
	}
	msgsBody_bs, _ := json.Marshal(msgsBody)
	gpp = &utils.GPP{
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
