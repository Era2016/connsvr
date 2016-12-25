package controller

import (
	"encoding/json"
	"net/http"
)

func (demo *Demo) Msgs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{
			"MsgId": "1",
			"Uid":   "",
			"Sid":   "",
			"Body":  "old_msg_1",
		},
		{
			"MsgId": "2",
			"Uid":   "",
			"Sid":   "",
			"Body":  "old_msg_2",
		},
	})
}
