package httpapi

import (
	"encoding/json"
	"net/http"
)

// Result mirrors Java CommonResult: {code, msg, data}, code 0 = success.
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func WriteOK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Result{Code: 0, Msg: "ok", Data: data})
}

func WriteErr(w http.ResponseWriter, httpStatus int, code int, msg string) {
	write(w, httpStatus, Result{Code: code, Msg: msg, Data: nil})
}

func write(w http.ResponseWriter, httpStatus int, body Result) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(body)
}
