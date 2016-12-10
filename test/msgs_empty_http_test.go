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

func TestMsgsEmptyHttp(t *testing.T) {
	subcmd := "1"
	rid := "r1"
	uid := "u_TestMsgsEmptyHttp"

	gpp := &utils.GPP{
		Uri: fmt.Sprintf("http://:%d", conf.C.App.Hport),
		Headers: map[string]string{
			"Connection": "Close",
		},
		Params: map[string]string{
			"cmd":      strconv.Itoa(int(comm.MSGS)),
			"rid":      rid,
			"uid":      uid,
			"subcmd":   subcmd,
			"callback": "",
		},
	}
	resp, err := utils.Get(gpp)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]string
	json.Unmarshal(resp, &m)

	t.Log("get resp:", m["body"])
}
