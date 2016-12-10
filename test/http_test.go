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
	"sync"
	"testing"
	"time"
)

func TestHttp(t *testing.T) {
	cmd := comm.PUSH
	rid := "r1"
	uid := "u_TestHttp"
	sid := "s1"
	text := "hello world"

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		gpp := &utils.GPP{
			Uri: fmt.Sprintf("http://:%d", conf.C.App.Hport),
			Headers: map[string]string{
				"Connection": "Close",
			},
			Params: map[string]string{
				"cmd":      strconv.Itoa(int(comm.ENTER)),
				"rid":      rid,
				"uid":      uid,
				"sid":      sid,
				"callback": "",
			},
		}
		resp, err := utils.Get(gpp)
		if err != nil {
			t.Fatal(err)
		}

		var m map[string]string
		json.Unmarshal(resp, &m)
		if _cmd := m["cmd"]; _cmd != strconv.Itoa(int(cmd)) {
			t.Errorf("get: %v, expected: %v", _cmd, cmd)
		}
		if _uid := m["uid"]; _uid != uid {
			t.Errorf("get: %s, expected: %s", _uid, uid)
		}
		if _rid := m["rid"]; _rid != rid {
			t.Errorf("get: %s, expected: %s", _rid, rid)
		}
		if body := m["body"]; body != text {
			t.Errorf("get: %s, expected: %s", body, text)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 50)

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
		msg.SetRid(rid)
		msg.SetBody(text)
		data, ok := msg.Encode()
		if !ok {
			t.Fatal("msg.Encode() error")
		}

		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()
}
