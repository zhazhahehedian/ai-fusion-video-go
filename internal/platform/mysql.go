package platform

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/config"
)

// WithUTCTimeZone 让每个连接执行 SET time_zone='+00:00'。
// 内容表仍用 CURRENT_TIMESTAMP，但墙钟成为 UTC，和 DSN 的 loc=UTC 标注一致。
// 调用方已经写了 time_zone 时保留原值。
func WithUTCTimeZone(dsn string) string {
	if strings.Contains(dsn, "time_zone=") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&time_zone=%27%2B00%3A00%27"
	}
	return dsn + "?time_zone=%27%2B00%3A00%27"
}

func OpenMySQL(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", WithUTCTimeZone(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL: %w", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("MySQL 探活失败: %w", err)
	}
	return db, nil
}
