package handler

import "net/http"

// HTTPInterceptor 请求拦截器，鉴权
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseMultipartForm(32 << 20)
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			if len(username) < 3 && !ISTokenValid(token) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		})
}
