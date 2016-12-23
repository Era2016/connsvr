package controller

import (
	"encoding/json"
	"net/http"

	"github.com/simplejia/connsvr/test/logicsvr/service"
)

func (demo *Demo) Pub(w http.ResponseWriter, r *http.Request) {
	rid := r.FormValue("rid")
	body := r.FormValue("body")
	demoService := service.NewDemo()
	demoService.Pub(rid, body)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
	})
}
