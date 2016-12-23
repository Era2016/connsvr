package filter

import (
	"net/http"
	"strings"
)

func Method(w http.ResponseWriter, r *http.Request, p map[string]interface{}) bool {
	method, ok := p["type"].(string)
	if ok && strings.ToLower(r.Method) != strings.ToLower(method) {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
