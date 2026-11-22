# 媒体进程、日志与错误

## 解决什么问题

ffmpeg 可能跑几分钟。取消 HTTP 或租约失效时，进程必须被杀掉，不能变成孤儿。日志要能把同一次 HTTP 请求的访问行、鉴权失败和下游错误串起来。接口错误要稳定成 `{code,msg,data}`，而不是把 Go 的 error 文本直接扔给前端。

## 怎么实现

`internal/media/process.go` 的 `ProcessRunner` 用 `exec.CommandContext`。额外的 `onTick` 每秒执行一次，媒体 worker 用它核对租约。`ctx` 取消或 `onTick` 返回错误时调用 `Kill`。超时默认 30 分钟。

```36:47:internal/media/process.go
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workDir
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("无法启动进程 %q: %w", argv[0], err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
```

单测覆盖取消和 `onTick` 杀进程，不依赖本机安装 ffmpeg。worker 只有在 `FFMPEG_VERSION` 这个 kind 下才真正调用 ffmpeg，并且只执行 `-version`。

日志是 `log/slog`。`internal/logging/logging.go` 按配置创建 text 或 json handler。访问日志中间件 `internal/httpapi/middleware_log.go` 读取或生成 `X-Request-ID`，放进 logger 和 context：

```21:36:internal/httpapi/middleware_log.go
			id := incomingRequestID(r.Header.Get("X-Request-ID"))
			if id == "" {
				generated, err := logging.NewRequestID()
				// ...
				id = generated
			}
			w.Header().Set("X-Request-ID", id)
			logger := base.With(slog.String("request_id", id))
			ctx := logging.WithRequestID(logging.NewContext(r.Context(), logger), id)
```

请求 ID 只接受最长 128 的 ASCII。非法的请求头会被忽略并重新生成，避免把换行写进日志。logger 放在一个可替换的盒子里，后面的中间件可以追加字段，已经发出去的 context 仍看得到。

业务错误是 `internal/apperr.Error`：HTTP 状态、业务 `code`、中文 `msg`。`WriteError` 用 `errors.As` 识别它。未识别的错误记一条带 `request_id` 的 Error 日志，对外只返回“服务器内部错误”。成功体是 `code = 0`，`msg` 是 `success`，与 Java `CommonResult` 一致。

领域代码返回 `apperr.BadRequest`、`Forbidden`、`NotFound`、`Conflict`，不在深层 `log.Fatal`，也不直接写 `http.ResponseWriter`。

## 和 Java 的对应与差异

Java 的媒体执行器同样在命令运行期间做租约检查，取消时销毁进程。Go 用 `CommandContext` 加上显式 `Kill`。Java 常见的是 SLF4J MDC；Go 把 `request_id` 放在 slog 属性里，没有全局 MDC。

`CommonResult` 与 `Result` 都是 `code`、`msg`、`data`。分页列表是 `list` 与 `total`，查询参数是 `pageNo`、`pageSize`。

## 常见坑

- 只靠 `CommandContext`，不在 `onTick` 失败时 `Kill`。上下文取消和租约失败是两条路径，现有 Runner 两条都会杀进程。
- 把请求体或 token 打进访问日志。现有访问行只有方法、路径、状态码和耗时。
- 在 SQL 函数里 `fmt.Errorf` 后被 handler 当成 500。用户能看懂的校验失败要返回 `*apperr.Error`。
