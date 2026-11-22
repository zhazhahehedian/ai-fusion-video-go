package handlers

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func assetProject(row sqlcdb.AfvAsset) (int64, error) {
	if !row.ProjectID.Valid || row.ProjectID.Int64 <= 0 {
		return 0, apperr.BadRequest("资产缺少所属项目")
	}
	return row.ProjectID.Int64, nil
}

type AssetCreate struct{}

func (AssetCreate) ResourceType() resource.Type   { return resource.TypeAsset }
func (AssetCreate) Operation() resource.Operation { return resource.OpCreate }

func (AssetCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvAsset{
		UserID:     sql.NullInt64{Int64: actorUserID, Valid: true},
		ProjectID:  sql.NullInt64{Int64: projectID, Valid: true},
		OwnerType:  sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:    sql.NullInt64{Int64: actorUserID, Valid: true},
		SourceType: sql.NullInt32{Int32: 1, Valid: true},
		Status:     sql.NullInt32{Int32: 1, Valid: true},
	}
	if _, err := applyAssetFields(&row, create.Fields, true); err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateAsset(ctx, sqlcdb.CreateAssetParams{
		UserID:      row.UserID,
		ProjectID:   row.ProjectID,
		Type:        row.Type,
		Name:        row.Name,
		Description: row.Description,
		CoverUrl:    row.CoverUrl,
		Properties:  row.Properties,
		Tags:        row.Tags,
		SourceType:  row.SourceType,
		AiPrompt:    row.AiPrompt,
		OwnerID:     row.OwnerID,
		Status:      row.Status,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetAssetByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	return created(resource.TypeAsset, id, projectID, resource.AssetVO(saved))
}

type AssetPatch struct{}

func (AssetPatch) ResourceType() resource.Type   { return resource.TypeAsset }
func (AssetPatch) Operation() resource.Operation { return resource.OpPatch }

func (AssetPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetAssetByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "资产不存在")
	}
	projectID, err := assetProject(row)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row, err := q.LockAssetByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "资产不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("资产版本已变化，请刷新后重试")
	}
	before := resource.AssetVO(row)
	changed, err := applyAssetFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateAsset(ctx, sqlcdb.UpdateAssetParams{
		Type:        row.Type,
		Name:        row.Name,
		Description: row.Description,
		CoverUrl:    row.CoverUrl,
		Properties:  row.Properties,
		Tags:        row.Tags,
		SourceType:  row.SourceType,
		AiPrompt:    row.AiPrompt,
		Status:      row.Status,
		ID:          row.ID,
		Version:     row.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetAssetByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := assetProject(saved)
	if err != nil {
		return resource.Mutation{}, err
	}
	return patched(resource.TypeAsset, saved.ID, projectID, saved.Version, changed, before, resource.AssetVO(saved))
}

type AssetDelete struct{}

func (AssetDelete) ResourceType() resource.Type   { return resource.TypeAsset }
func (AssetDelete) Operation() resource.Operation { return resource.OpDelete }

func (AssetDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetAssetByID(ctx, del.ID)
	if err != nil {
		return nil, missing(err, "资产不存在")
	}
	projectID, err := assetProject(row)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row, err := q.LockAssetByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "资产不存在")
	}
	if row.Version != del.Version {
		return resource.Mutation{}, apperr.Conflict("资产版本已变化，请刷新后重试")
	}
	projectID, err := assetProject(row)
	if err != nil {
		return resource.Mutation{}, err
	}
	items, err := q.LockAssetItemsByAsset(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	changes := make([]resource.Change, 0, len(items)+1)
	for _, item := range items {
		n, err := q.SoftDeleteAssetItem(ctx, sqlcdb.SoftDeleteAssetItemParams{ID: item.ID, Version: item.Version})
		if err != nil {
			return resource.Mutation{}, err
		}
		if err := oneRow(n, "子资产版本已变化，请刷新后重试"); err != nil {
			return resource.Mutation{}, err
		}
		itemBefore := resource.AssetItemVO(item)
		change, err := deleteChange(resource.TypeAssetItem, item.ID, projectID, item.Version+1, itemBefore, markDeleted(itemBefore, item.Version+1), []string{"deleted"})
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
	}
	n, err := q.SoftDeleteAsset(ctx, sqlcdb.SoftDeleteAssetParams{ID: row.ID, Version: row.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	before := resource.AssetVO(row)
	change, err := deleteChange(resource.TypeAsset, row.ID, projectID, row.Version+1, before, markDeleted(before, row.Version+1), []string{"deleted"})
	if err != nil {
		return resource.Mutation{}, err
	}
	changes = append(changes, change)
	return resource.Mutation{Result: true, Changes: changes, PreviousValues: before}, nil
}

type AssetItemCreate struct{}

func (AssetItemCreate) ResourceType() resource.Type   { return resource.TypeAssetItem }
func (AssetItemCreate) Operation() resource.Operation { return resource.OpCreate }

func (AssetItemCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	assetID, err := requiredPositiveID(create.Fields["assetId"], "assetId")
	if err != nil {
		return nil, err
	}
	row, err := q.GetAssetByID(ctx, assetID)
	if err != nil {
		return nil, missing(err, "资产不存在")
	}
	projectID, err := assetProject(row)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetItemCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	assetID, err := requiredPositiveID(create.Fields["assetId"], "assetId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	asset, err := q.LockAssetByID(ctx, assetID)
	if err != nil {
		return resource.Mutation{}, missing(err, "资产不存在")
	}
	if asset.Version != parentVersion {
		return resource.Mutation{}, apperr.Conflict("资产版本已变化，请刷新后重试")
	}
	fields := cloneMap(create.Fields)
	delete(fields, "assetId")
	delete(fields, "parentVersion")
	count, err := q.CountAssetItems(ctx, asset.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvAssetItem{
		AssetID:    asset.ID,
		SortOrder:  sql.NullInt32{Int32: int32(count), Valid: true},
		SourceType: sql.NullInt32{Int32: 1, Valid: true},
	}
	if _, err := applyAssetItemFields(&row, fields, true); err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateAssetItem(ctx, sqlcdb.CreateAssetItemParams{
		AssetID:      row.AssetID,
		ItemType:     row.ItemType,
		Name:         row.Name,
		ImageUrl:     row.ImageUrl,
		ThumbnailUrl: row.ThumbnailUrl,
		Properties:   row.Properties,
		SortOrder:    row.SortOrder,
		SourceType:   row.SourceType,
		AiPrompt:     row.AiPrompt,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpAssetVersion(ctx, sqlcdb.BumpAssetVersionParams{ID: asset.ID, Version: asset.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetAssetItemByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := assetProject(asset)
	if err != nil {
		return resource.Mutation{}, err
	}
	return created(resource.TypeAssetItem, id, projectID, resource.AssetItemVO(saved))
}

type AssetItemPatch struct{}

func (AssetItemPatch) ResourceType() resource.Type   { return resource.TypeAssetItem }
func (AssetItemPatch) Operation() resource.Operation { return resource.OpPatch }

func (AssetItemPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	item, err := q.GetAssetItemByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "子资产不存在")
	}
	asset, err := q.GetAssetByID(ctx, item.AssetID)
	if err != nil {
		return nil, missing(err, "资产不存在")
	}
	projectID, err := assetProject(asset)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetItemPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	if patch.ParentVersion == nil {
		return resource.Mutation{}, apperr.BadRequest("assetVersion 无效")
	}
	current, err := q.GetAssetItemByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "子资产不存在")
	}
	asset, err := q.LockAssetByID(ctx, current.AssetID)
	if err != nil {
		return resource.Mutation{}, missing(err, "资产不存在")
	}
	if asset.Version != *patch.ParentVersion {
		return resource.Mutation{}, apperr.Conflict("资产版本已变化，请刷新后重试")
	}
	row, err := q.LockAssetItemByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "子资产不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("子资产版本已变化，请刷新后重试")
	}
	before := resource.AssetItemVO(row)
	changed, err := applyAssetItemFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateAssetItem(ctx, sqlcdb.UpdateAssetItemParams{
		ItemType:     row.ItemType,
		Name:         row.Name,
		ImageUrl:     row.ImageUrl,
		ThumbnailUrl: row.ThumbnailUrl,
		Properties:   row.Properties,
		SortOrder:    row.SortOrder,
		SourceType:   row.SourceType,
		AiPrompt:     row.AiPrompt,
		ID:           row.ID,
		Version:      row.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "子资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpAssetVersion(ctx, sqlcdb.BumpAssetVersionParams{ID: asset.ID, Version: asset.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetAssetItemByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := assetProject(asset)
	if err != nil {
		return resource.Mutation{}, err
	}
	return patched(resource.TypeAssetItem, saved.ID, projectID, saved.Version, changed, before, resource.AssetItemVO(saved))
}

type AssetItemDelete struct{}

func (AssetItemDelete) ResourceType() resource.Type   { return resource.TypeAssetItem }
func (AssetItemDelete) Operation() resource.Operation { return resource.OpDelete }

func (AssetItemDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return nil, err
	}
	item, err := q.GetAssetItemByID(ctx, del.ID)
	if err != nil {
		return nil, missing(err, "子资产不存在")
	}
	asset, err := q.GetAssetByID(ctx, item.AssetID)
	if err != nil {
		return nil, missing(err, "资产不存在")
	}
	projectID, err := assetProject(asset)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (AssetItemDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	if del.ParentVersion == nil {
		return resource.Mutation{}, apperr.BadRequest("parentVersion 无效")
	}
	current, err := q.GetAssetItemByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "子资产不存在")
	}
	asset, err := q.LockAssetByID(ctx, current.AssetID)
	if err != nil {
		return resource.Mutation{}, missing(err, "资产不存在")
	}
	if asset.Version != *del.ParentVersion {
		return resource.Mutation{}, apperr.Conflict("资产版本已变化，请刷新后重试")
	}
	row, err := q.LockAssetItemByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "子资产不存在")
	}
	if row.Version != del.Version {
		return resource.Mutation{}, apperr.Conflict("子资产版本已变化，请刷新后重试")
	}
	n, err := q.SoftDeleteAssetItem(ctx, sqlcdb.SoftDeleteAssetItemParams{ID: row.ID, Version: row.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "子资产版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := assetProject(asset)
	if err != nil {
		return resource.Mutation{}, err
	}
	assetBefore := resource.AssetVO(asset)
	assetAfter := cloneMap(assetBefore)
	assetAfter["version"] = asset.Version + 1
	changed := []string{"items"}
	if coverMatches(row.ImageUrl, asset.CoverUrl) {
		cleared, err := q.ClearAssetCover(ctx, sqlcdb.ClearAssetCoverParams{ID: asset.ID, Version: asset.Version})
		if err != nil {
			return resource.Mutation{}, err
		}
		if err := oneRow(cleared, "资产版本已变化，请刷新后重试"); err != nil {
			return resource.Mutation{}, err
		}
		assetAfter["coverUrl"] = nil
		changed = []string{"coverUrl"}
	} else {
		bumped, err := q.BumpAssetVersion(ctx, sqlcdb.BumpAssetVersionParams{ID: asset.ID, Version: asset.Version})
		if err != nil {
			return resource.Mutation{}, err
		}
		if err := oneRow(bumped, "资产版本已变化，请刷新后重试"); err != nil {
			return resource.Mutation{}, err
		}
	}
	parentChange, err := assetChange(asset.ID, projectID, asset.Version+1, changed, assetBefore, assetAfter)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.AssetItemVO(row)
	itemChange, err := deleteChange(resource.TypeAssetItem, row.ID, projectID, row.Version+1, before, markDeleted(before, row.Version+1), []string{"deleted"})
	if err != nil {
		return resource.Mutation{}, err
	}
	return resource.Mutation{
		Result:         true,
		Changes:        []resource.Change{itemChange, parentChange},
		PreviousValues: before,
	}, nil
}
