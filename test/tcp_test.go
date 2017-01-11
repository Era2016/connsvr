package tests

import (
	"bufio"
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

func TestTcp(t *testing.T) {
	rid := "r1"
	uid := "u_TestTcp"
	sid := "s1"
	text := "hello world"

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

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
		msg.SetSid(sid)
		msg.SetRid(rid)
		msg.SetBody("")
		ok := msg.Encode(conn, nil)
		if !ok {
			t.Fatal("msg.Encode() error")
		}

		_msg := new(proto.MsgTcp)
		ok = _msg.Decode(bufio.NewReader(conn), nil, nil)
		if !ok {
			t.Fatal("_msg.DecodeHeader() error")
		}

		switch _msg.Cmd() {
		case comm.PUSH:
			if _msg.Uid() != uid {
				t.Errorf("get: %s, expected: %s", _msg.Uid(), uid)
			}
			if _msg.Rid() != rid {
				t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
			}
			if _msg.Body() != text {
				t.Errorf("get: %s, expected: %s", _msg.Body(), text)
			}
		case comm.MSGS:
			if _msg.Uid() != uid {
				t.Errorf("get: %s, expected: %s", _msg.Uid(), uid)
			}
			if _msg.Rid() != rid {
				t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
			}
			t.Log("get resp:", _msg.Body())
		default:
			t.Errorf("get: %v, not expected", _msg.Cmd())
		}
	}()

	time.Sleep(time.Millisecond * 30)

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
