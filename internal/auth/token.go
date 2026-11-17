package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	accessTokenPrefix   = "fusion:token:"
	refreshTokenPrefix  = "fusion:refresh_token:"
	accessRefreshSuffix = ":refresh"
	refreshAccessSuffix = ":access"

	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// TokenPair mirrors Java TokenService.TokenPair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// TokenStore manages Redis access/refresh tokens.
type TokenStore struct {
	rdb *redis.Client
}

func NewTokenStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

func (t *TokenStore) Create(ctx context.Context, session Session) (TokenPair, error) {
	raw, err := json.Marshal(session)
	if err != nil {
		return TokenPair{}, err
	}
	accessToken, err := randomToken()
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return TokenPair{}, err
	}

	pipe := t.rdb.Pipeline()
	pipe.Set(ctx, accessTokenPrefix+accessToken, raw, accessTokenTTL)
	pipe.Set(ctx, refreshTokenPrefix+refreshToken, raw, refreshTokenTTL)
	pipe.Set(ctx, accessRefreshKey(accessToken), refreshToken, refreshTokenTTL)
	pipe.Set(ctx, refreshAccessKey(refreshToken), accessToken, refreshTokenTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL / time.Second),
	}, nil
}

func (t *TokenStore) GetAccessSession(ctx context.Context, accessToken string) (Session, bool, error) {
	raw, err := t.rdb.Get(ctx, accessTokenPrefix+accessToken).Result()
	if err == redis.Nil {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	session, ok := deserializeSession(raw)
	return session, ok, nil
}

// PeekRefreshSession reads refresh session without consuming it (Java pre-check before rotate).
func (t *TokenStore) PeekRefreshSession(ctx context.Context, refreshToken string) (Session, bool, error) {
	raw, err := t.rdb.Get(ctx, refreshTokenPrefix+refreshToken).Result()
	if err == redis.Nil {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	session, ok := deserializeSession(raw)
	return session, ok, nil
}

// Refresh atomically consumes the refresh token (GETDEL) and issues a new pair.
func (t *TokenStore) Refresh(ctx context.Context, refreshToken string) (TokenPair, bool, error) {
	raw, err := t.rdb.GetDel(ctx, refreshTokenPrefix+refreshToken).Result()
	if err == redis.Nil {
		return TokenPair{}, false, nil
	}
	if err != nil {
		return TokenPair{}, false, err
	}
	session, ok := deserializeSession(raw)
	if !ok {
		return TokenPair{}, false, nil
	}

	oldAccess, _ := t.rdb.Get(ctx, refreshAccessKey(refreshToken)).Result()
	if oldAccess != "" {
		_ = t.rdb.Del(ctx,
			accessTokenPrefix+oldAccess,
			accessRefreshKey(oldAccess),
		).Err()
	}
	_ = t.rdb.Del(ctx, refreshAccessKey(refreshToken)).Err()

	pair, err := t.Create(ctx, session)
	if err != nil {
		return TokenPair{}, false, err
	}
	return pair, true, nil
}

func (t *TokenStore) RemoveAccess(ctx context.Context, accessToken string) error {
	refreshToken, err := t.rdb.Get(ctx, accessRefreshKey(accessToken)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	keys := []string{accessTokenPrefix + accessToken, accessRefreshKey(accessToken)}
	if refreshToken != "" {
		keys = append(keys, refreshTokenPrefix+refreshToken, refreshAccessKey(refreshToken))
	}
	return t.rdb.Del(ctx, keys...).Err()
}

func accessRefreshKey(accessToken string) string {
	return accessTokenPrefix + accessToken + accessRefreshSuffix
}

func refreshAccessKey(refreshToken string) string {
	return refreshTokenPrefix + refreshToken + refreshAccessSuffix
}

func randomToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func deserializeSession(raw string) (Session, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Session{}, false
	}
	var session Session
	if err := json.Unmarshal([]byte(raw), &session); err == nil && session.UserID != 0 {
		return session, true
	}
	// Legacy "userId:username"
	parts := strings.SplitN(raw, ":", 2)
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || userID == 0 {
		return Session{}, false
	}
	username := ""
	if len(parts) > 1 {
		username = parts[1]
	}
	return Session{UserID: userID, Username: username}, true
}
