package httpapi

import "net/http"

const DefaultMaxBodyBytes int64 = 8 << 20

func maxBodyBytes(deps Deps) int64 {
	if deps.MaxBodyBytes > 0 {
		return deps.MaxBodyBytes
	}
	return DefaultMaxBodyBytes
}

// LimitRequestBody 拒绝超过上限的请求体。脚本原文可能较长，默认 8MiB。
func LimitRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && maxBytes > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
