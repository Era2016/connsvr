package tests

import (
	"bufio"
	"encoding/json"
	"fmt"

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
	cmd := comm.PUSH
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
		data, ok := msg.Encode()
		if !ok {
			t.Fatal("msg.Encode() error")
		}

		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}

		_msg := new(proto.MsgTcp)
		ok = _msg.Decode(bufio.NewReader(conn))
		if !ok {
			t.Fatal("_msg.DecodeHeader() error")
		}

		switch _msg.Cmd() {
		case cmd:
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

	time.Sleep(time.Millisecond * 10)

	wg.Add(1)
	go func() {
		defer wg.Done()

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
		msg.SetUid("")
		msg.SetSid("")
		msg.SetRid(rid)
		msg.SetBody(text)
		pushExt := &comm.PushExt{
			MsgId: "1",
		}
		ext_bs, _ := json.Marshal(pushExt)
		msg.SetExt(string(ext_bs))
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
