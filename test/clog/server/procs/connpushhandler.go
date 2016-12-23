package procs

import (
	"encoding/json"
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/connsvr/api"
	"github.com/simplejia/connsvr/test/clog"
)

func ConnPushHandler(cate, subcate, body string, params map[string]interface{}) {
	pushMsg := &api.PushMsg{}
	err := json.Unmarshal([]byte(body), pushMsg)
	if err != nil {
		clog.Error("ConnPushHandler() json.Unmarshal body: %s, error: %v", body, err)
		return
	}

	var connRedisAddr *ConnRedisAddr
	bs, _ := json.Marshal(params["redis"])
	json.Unmarshal(bs, &connRedisAddr)
	if connRedisAddr == nil {
		log.Printf("ConnReportHandler() params not right: %v\n", params)
		return
	}

	addr, err := ConnRedisAddrFunc(connRedisAddr.AddrType, connRedisAddr.Addr)
	if err != nil {
		clog.Error("ConnPushHandler() ConnRedisAddrFunc error: %v", err)
		return
	}

	c, err := redis.Dial("tcp", addr)
	if err != nil {
		clog.Error("ConnPushHandler() redis.Dial error: %v", err)
		return
	}

	ips, err := redis.Strings(c.Do("ZRANGE", "conn:ips", 0, -1))
	if err != nil {
		clog.Error("ConnPushHandler() redis get ips error: %v", err)
		return
	}

	clog.Info("ConnPushHandler() ips: %v, pushMsg: %+v", ips, pushMsg)

	api.PushRaw(pushMsg, ips)

	return
}

func init() {
	RegisterHandler("connpushhandler", ConnPushHandler)
}
