package clog

import (
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/test/logicsvr/conf"
)

func init() {
	clog.Init(conf.C.Clog.Name, "", conf.C.Clog.Level, conf.C.Clog.Mode)
}
