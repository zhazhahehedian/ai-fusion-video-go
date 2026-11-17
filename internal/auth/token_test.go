package auth

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRefreshConsumesRefreshTokenOnce(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:46379", Password: "123456", DB: 15})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = rdb.FlushDB(context.Background()).Err()
		_ = rdb.Close()
	})

	store := NewTokenStore(rdb)
	pair, err := store.Create(ctx, Session{UserID: 42, Username: "alice"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	first, ok, err := store.Refresh(ctx, pair.RefreshToken)
	if err != nil || !ok {
		t.Fatalf("first refresh: ok=%v err=%v", ok, err)
	}
	if first.AccessToken == "" || first.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected rotated tokens, got %+v", first)
	}

	_, ok, err = store.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("second refresh err: %v", err)
	}
	if ok {
		t.Fatal("expected consumed refresh token to fail second use")
	}
}
