# 分镜四层

分镜是项目下的镜头工作区，和剧本树分开。

| 层 | 资源类型 | 父层 |
|---|---|---|
| 分镜 | `storyboard` | 项目 |
| 分镜集 | `storyboardEpisode` | 分镜 |
| 分镜场次 | `storyboardScene` | 分镜集 |
| 镜头 | `shot` | 分镜场次 |

类型字符串用 Java 的驼峰：`storyboardEpisode`、`storyboardScene`、`shot`。剧本树仍是 `script_episode`、`script_scene`，避免改掉已经写入的修订类型。

## 为什么没有删除分镜文档

Java `StoryboardResourceDeleteHandler` 只处理分镜集、场次和镜头。分镜文档本身没有删除接口。清空用的是 `CLEAR`：子层软删，分镜行还在。Go 因此不提供 `DELETE /api/storyboard/{id}`。

`POST /api/storyboard/{id}/clearContent` 在一个事务里锁定分镜、分镜集、场次和镜头，然后按镜头、场次、分镜集的顺序软删。分镜集写入 `deleted_id = id`，否则唯一键会挡住同一剧本分集再次挂上来。分镜版本加一，返回清空后的分镜。

`POST /api/storyboard/episode/{id}/clearContent` 只清这一集下面的场次和镜头，分镜集保留并加版本。

## 删除子层

删除分镜集：镜头、场次、分镜集，顺序固定，每行一条 revision。删除场次：先删镜头，再删场次。删除镜头：只删自己。

创建子层要带父版本。父版本不对就 409，调用方刷新后再提交。

## 尚未实现

批量维护子层、合成视频、统计和配乐回写。这些要么是另一套批量命令，要么要媒体作业写回 `bgm_*` 和 `compose_job_id`。列已经在表上，没有对应的写接口。
