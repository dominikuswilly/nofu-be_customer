package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // intentionally ignore encode error; caller can add logging
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiResp{
		ResponseCode:    fmt.Sprintf("%d", status),
		ResponseMessage: msg,
		Data:            nil,
	})
}
