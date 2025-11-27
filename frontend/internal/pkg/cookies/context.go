package cookies

import (
	"context"
	"net/http"
)

type cookiesKey struct{}

func WithCookies(ctx context.Context, cookies []*http.Cookie) context.Context {
	return context.WithValue(ctx, cookiesKey{}, cookies)
}

func FromContext(ctx context.Context) []*http.Cookie {
	cookies, _ := ctx.Value(cookiesKey{}).([]*http.Cookie)
	return cookies
}
