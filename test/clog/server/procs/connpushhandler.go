package procs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/connsvr/api"
	"github.com/simplejia/connsvr/test/clog"
	"github.com/simplejia/lm"
	"github.com/simplejia/utils"
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

	ext := ""
	if pushMsg.Ext != nil {
		bs, _ := json.Marshal(pushMsg.Ext)
		ext = string(bs)
	}

	gpp := &utils.GPP{
		Params: map[string]string{
			"cmd":    strconv.Itoa(int(pushMsg.Cmd)),
			"subcmd": strconv.Itoa(int(pushMsg.Subcmd)),
			"rid":    pushMsg.Rid,
			"uid":    pushMsg.Uid,
			"sid":    pushMsg.Sid,
			"body":   pushMsg.Body,
			"ext":    ext,
		},
		Timeout: time.Millisecond * 50,
	}
	for _, ipport := range ips {
		gpp.Uri = fmt.Sprintf("http://%s", ipport)
		_, err := utils.Post(gpp)
		if err != nil {
			clog.Error("ConnPushHandler() utils.Post error, gpp: %+v, ipport: %s", gpp, ipport)
			continue
		}
	}

	return
}

func init() {
	RegisterHandler("connpushhandler", ConnPushHandler)
}
