package controller

import (
	"encoding/json"
	"net/http"
)

func (demo *Demo) Var(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PushKind":         2,
		"RoomWithPushKind": nil,
		"MsgNum":           20,
	})
}
