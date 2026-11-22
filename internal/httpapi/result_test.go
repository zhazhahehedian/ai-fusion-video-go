package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteOKMessageIsSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteOK(rec, map[string]int{"id": 1})
	if !strings.Contains(rec.Body.String(), `"msg":"success"`) {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestLimitRequestBody(t *testing.T) {
	handler := LimitRequestBody(16)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 64)
		_, err := r.Body.Read(buf)
		if err == nil {
			WriteOK(w, nil)
			return
		}
		WriteErr(w, http.StatusBadRequest, 400, "请求体无效")
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"too-long-for-the-cap"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
