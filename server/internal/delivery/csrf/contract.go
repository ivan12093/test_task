package csrf

import "context"

type CSRFTokenUC interface {
	GetCSRFToken(ctx context.Context) (string, error)
	ValidateCSRFToken(ctx context.Context, tokenCookie, tokenHeader string) (bool, error)
}
