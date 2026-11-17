package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type Service struct {
	q *sqlcdb.Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{q: sqlcdb.New(db)}
}

type PageResult struct {
	List  []map[string]any `json:"list"`
	Total int64            `json:"total"`
}

func (s *Service) Page(ctx context.Context, ownerUserID int64, pageNo, pageSize int) (PageResult, error) {
	pageNo, pageSize, err := normalizePage(pageNo, pageSize)
	if err != nil {
		return PageResult{}, err
	}
	total, err := s.q.CountOwnedProjects(ctx, ownerUserID)
	if err != nil {
		return PageResult{}, err
	}
	offset := int32((pageNo - 1) * pageSize)
	rows, err := s.q.ListOwnedProjects(ctx, sqlcdb.ListOwnedProjectsParams{
		OwnerID: ownerUserID,
		Limit:   int32(pageSize),
		Offset:  offset,
	})
	if err != nil {
		return PageResult{}, err
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, toVO(row.ID, row.Name, row.Description, row.CoverUrl, row.Scope, row.OwnerType, row.OwnerID, row.Status, row.Properties, row.ArtStyle, row.ArtStyleDescription, row.ArtStyleImagePrompt, row.ArtStyleImageUrl, row.Version, row.Deleted, row.CreateTime, row.UpdateTime))
	}
	return PageResult{List: list, Total: total}, nil
}

func (s *Service) List(ctx context.Context, ownerUserID int64) ([]map[string]any, error) {
	rows, err := s.q.ListAllOwnedProjects(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, toVO(row.ID, row.Name, row.Description, row.CoverUrl, row.Scope, row.OwnerType, row.OwnerID, row.Status, row.Properties, row.ArtStyle, row.ArtStyleDescription, row.ArtStyleImagePrompt, row.ArtStyleImageUrl, row.Version, row.Deleted, row.CreateTime, row.UpdateTime))
	}
	return list, nil
}

func normalizePage(pageNo, pageSize int) (int, int, error) {
	if pageNo < 1 {
		return 0, 0, apperr.BadRequest("页码最小为 1")
	}
	if pageSize < 1 {
		return 0, 0, apperr.BadRequest("每页条数最小为 1")
	}
	if pageSize > 100 {
		return 0, 0, apperr.BadRequest("每页条数最大为 100")
	}
	return pageNo, pageSize, nil
}

func toVO(
	id int64,
	name string,
	description sql.NullString,
	coverUrl sql.NullString,
	scope sql.NullInt32,
	ownerType int32,
	ownerID int64,
	status sql.NullInt32,
	properties json.RawMessage,
	artStyle sql.NullString,
	artStyleDescription sql.NullString,
	artStyleImagePrompt sql.NullString,
	artStyleImageUrl sql.NullString,
	version int32,
	deleted int8,
	createTime time.Time,
	updateTime time.Time,
) map[string]any {
	return map[string]any{
		"id":                  id,
		"name":                name,
		"description":         nullString(description),
		"coverUrl":            nullString(coverUrl),
		"scope":               nullInt32(scope),
		"ownerType":           ownerType,
		"ownerId":             ownerID,
		"status":              nullInt32(status),
		"properties":          rawJSON(properties),
		"artStyle":            nullString(artStyle),
		"artStyleDescription": nullString(artStyleDescription),
		"artStyleImagePrompt": nullString(artStyleImagePrompt),
		"artStyleImageUrl":    nullString(artStyleImageUrl),
		"version":             version,
		"deleted":             deleted != 0,
		"createTime":          formatTime(createTime),
		"updateTime":          formatTime(updateTime),
	}
}

func nullString(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}

func nullInt32(v sql.NullInt32) any {
	if !v.Valid {
		return nil
	}
	return v.Int32
}

func rawJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(v, &decoded); err != nil {
		return nil
	}
	return decoded
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}
