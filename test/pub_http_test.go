package tests

import (
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/utils"

	"testing"
)

func TestHttpPub(t *testing.T) {
	rid := "r1"
	uid := "u_TestHttpPub"
	sid := "s1"
	text := "hello world"

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
		t.Logf("get: %v, expected: %v", _cmd, comm.PUB)
		t.Logf("please check you conf(pubs)!!!")
	}

	t.Log("get resp:", m["body"])
}
