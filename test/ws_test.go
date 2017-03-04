package test

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/utils"

	"sync"
	"testing"
	"time"
)

func TestWS(t *testing.T) {
	rid := "r1"
	uid := "u_TestWS"
	sid := "s1"
	text := "hello world"

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		uri := fmt.Sprintf(
			"ws://:%d?cmd=%d&subcmd=0&rid=%s&uid=%s&sid=%s&body=&callback=",
			conf.C.App.Wport,
			comm.ENTER,
			rid,
			uid,
			sid,
		)
		c, _, err := websocket.DefaultDialer.Dial(uri, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		_, resp, err := c.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}

		var m map[string]string
		json.Unmarshal(resp, &m)

		switch m["cmd"] {
		case strconv.Itoa(int(comm.PUSH)):
			if _uid := m["uid"]; _uid != uid {
				t.Errorf("get: %s, expected: %s", _uid, uid)
			}
			if _rid := m["rid"]; _rid != rid {
				t.Errorf("get: %s, expected: %s", _rid, rid)
			}
			if body := m["body"]; body != text {
				t.Errorf("get: %s, expected: %s", body, text)
			}
		case strconv.Itoa(int(comm.MSGS)):
			if _uid := m["uid"]; _uid != uid {
				t.Errorf("get: %s, expected: %s", _uid, uid)
			}
			if _rid := m["rid"]; _rid != rid {
				t.Errorf("get: %s, expected: %s", _rid, rid)
			}
			t.Log("get resp:", m["body"])
		default:
			t.Errorf("get: %v, not expected", m["cmd"])
		}
	}()

	time.Sleep(time.Millisecond * 50)

	wg.Add(1)
	go func() {
		defer wg.Done()

		pushExt := &comm.PushExt{
			MsgId: "1",
		}
		ext_bs, _ := json.Marshal(pushExt)

		gpp := &utils.GPP{
			Uri: fmt.Sprintf("http://%s:%d", utils.LocalIp, conf.C.App.Bport),
			Params: map[string]string{
				"cmd":  strconv.Itoa(int(comm.PUSH)),
				"rid":  rid,
				"uid":  "",
				"sid":  "",
				"body": text,
				"ext":  string(ext_bs),
			},
		}
		_, err := utils.Post(gpp)
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()
}
