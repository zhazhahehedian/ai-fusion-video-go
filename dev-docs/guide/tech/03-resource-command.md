# 资源命令

## 要解决什么

创建、修改、删除和少数结构操作必须幂等，并且每次写入留下可回放的修订。

## 在哪里

- 受理：`internal/resource/service.go`
- 哈希：`internal/resource/hash.go`
- 各资源的 Authorize / Apply：`internal/resource/handlers/`

同一 `(actor, requestId)` 只能有一条命令。请求体经 `HashCommand` 得到 SHA-256。相同哈希重放返回原来的回执；不同内容返回 409。Authorize 与 Apply 在同一个事务里，先锁再写。

每条 `resource.Change` 写入 `afv_resource_revision`。`changedFields` 在提交时按字母序排列。map 的遍历顺序不稳定，不排序会导致同一命令的修订 JSON 在两次运行里不同。

分镜、镜头、资产和图片/视频任务上的 JSON 列允许为 NULL。读取语句用 `COALESCE(列, CAST('null' AS JSON))`，避免驱动把 SQL NULL 扫进 `json.RawMessage`。接口里这些字段仍然是 JSON null，不是空数组。

## 清空内容

`CLEAR_CONTENT` 只用于分镜和分镜集。它软删子层、给父行加版本，不删除父行。HTTP 是：

- `POST /api/storyboard/{id}/clearContent`
- `POST /api/storyboard/episode/{id}/clearContent`

请求体是 `{requestId, version}`，和前端 `storyboard.ts` 的 `clearContent` 一致。返回值是保留下来的父对象。

## 替换剧本原文

`ScriptSource.Apply` 不再用一条 `UPDATE` 把场次和分集全部标删除。它按 id 锁定每一行，逐行软删，并为每一行写 revision，然后再更新剧本。顺序是场次、分集、剧本。

## 和 Java 的对应

Java `ResourceCommandService.execute` 同样是一次事务里的授权、变更和修订。分镜清空对应 `StoryboardChildrenCommand.Mode.CLEAR`。Java 的批量场次、资产批量、分镜子层批量命令没有搬过来，见技术债。
