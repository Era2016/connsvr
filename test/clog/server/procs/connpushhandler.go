package procs

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/connsvr/api"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/connsvr/test/clog"
	"github.com/simplejia/lm"
)

func ConnPushHandler(cate, subcate, body string, params map[string]interface{}) {
	pushMsg := &api.PushMsg{}
	err := json.Unmarshal([]byte(body), pushMsg)
	if err != nil {
		clog.Error("ConnPushHandler() json.Unmarshal body: %s, error: %v", body, err)
		return
	}

	var ips []string

	lmStru := &lm.LmStru{
		Input:  CONN_IPS_KEY,
		Output: &ips,
		Proc: func(p, r interface{}) (err error) {
			defer func() {
				if err != nil {
					clog.Error("ConnPushHandler() get ips error: %v", err)
				}
			}()

			var connRedisAddr *ConnRedisAddr
			bs, _ := json.Marshal(params["redis"])
			json.Unmarshal(bs, &connRedisAddr)
			if connRedisAddr == nil {
				return fmt.Errorf("ConnReportHandler() params not right: %v", params)
			}

			addr, err := ConnRedisAddrFunc(connRedisAddr.AddrType, connRedisAddr.Addr)
			if err != nil {
				return fmt.Errorf("ConnPushHandler() ConnRedisAddrFunc error: %v", err)
			}

			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return fmt.Errorf("ConnPushHandler() redis.Dial error: %v", err)
			}

			ips, err := redis.Strings(c.Do("ZRANGE", p.(string), 0, -1))
			if err != nil {
				return fmt.Errorf("ConnPushHandler() redis get ips error: %v", err)
			}

			*r.(*[]string) = ips
			return nil
		},
		Key: func(p interface{}) string {
			return p.(string)
		},
		Lc: &lm.LcStru{
			Expire: time.Second * 30,
		},
	}

	err = lm.GlueLc(lmStru)
	if err != nil {
		clog.Error("ConnPushHandler() lm.GlueLc error: %v", err)
		return
	}

	clog.Info("ConnPushHandler() ips: %v, pushMsg: %+v", ips, pushMsg)

	msg := proto.NewMsg(comm.UDP)
	msg.SetCmd(pushMsg.Cmd)
	msg.SetSubcmd(pushMsg.Subcmd)
	msg.SetUid(pushMsg.Uid)
	msg.SetSid(pushMsg.Sid)
	msg.SetRid(pushMsg.Rid)
	msg.SetBody(pushMsg.Body)
	if pushMsg.Ext != nil {
		bs, _ := json.Marshal(pushMsg.Ext)
		msg.SetExt(string(bs))
	}
	data, ok := msg.Encode()
	if !ok {
		clog.Error("ConnPushHandler() msg encode error, msg: %+v", msg)
		return
	}

	for _, ipport := range ips {
		conn, err := net.Dial("udp", ipport)
		if err != nil {
			clog.Error("ConnPushHandler() dial ipport: %s, error: %v", ipport, err)
			continue
		}
		defer conn.Close()

		_, err = conn.Write(data)
		if err != nil {
			clog.Error("ConnPushHandler() conn.Write ipport: %s, error: %v", ipport, err)
			continue
		}
	}

	return
}

func init() {
	RegisterHandler("connpushhandler", ConnPushHandler)
}
