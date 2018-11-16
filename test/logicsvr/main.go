//go:generate wsp

package main

import (
	"fmt"
	"os"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/lc"

	"net/http"

	_ "github.com/simplejia/connsvr/test/logicsvr/clog"
	"github.com/simplejia/connsvr/test/logicsvr/conf"
	_ "github.com/simplejia/connsvr/test/logicsvr/mysql"
	_ "github.com/simplejia/connsvr/test/logicsvr/redis"
)

func init() {
	lc.Init(1e5)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

func main() {
	clog.Info("main()")

	s := &http.Server{
		Addr: fmt.Sprintf("%s:%d", conf.C.App.Ip, conf.C.App.Port),
	}
	err := s.ListenAndServe()
	clog.Error("main() s.ListenAndServe %v", err)
	os.Exit(-1)
}
