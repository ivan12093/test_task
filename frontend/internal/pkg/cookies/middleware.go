package cookies

import "net/http"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithCookies(r.Context(), r.Cookies())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
