package auth

import "context"

// Session is the Redis-backed login session (aligned with Java TokenService.TokenSession).
type Session struct {
	UserID        int64  `json:"userId"`
	Username      string `json:"username"`
	CurrentTeamID *int64 `json:"currentTeamId"`
}

type ctxKey int

const sessionKey ctxKey = 1

func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func SessionFrom(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}
