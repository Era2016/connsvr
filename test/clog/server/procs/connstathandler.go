package procs

import (
	"encoding/json"

	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/test/clog"
)

// body is comm.Stat
func ConnStatHandler(cate, subcate, body string, params map[string]interface{}) {
	stat := &comm.Stat{}
	err := json.Unmarshal([]byte(body), stat)
	if err != nil {
		clog.Error("ConnStatHandler() json.Unmarshal body: %s, error: %v", body, err)
		return
	}

	// TODO
	// 上报boss系统用于统计房间用户，推送耗时等
	clog.Info("ConnStatHandler() stat: %+v", stat)

	return
}

func init() {
	RegisterHandler("connstathandler", ConnStatHandler)
}
