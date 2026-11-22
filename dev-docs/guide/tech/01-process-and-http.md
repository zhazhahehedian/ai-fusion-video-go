# 进程与 HTTP

## 要解决什么

一个进程里同时跑 HTTP、发件箱、生成 worker 和媒体 worker。慢请求或大请求体不能占住连接；成功响应要让现有前端按 Java 的字段判断。

## 在哪里

- 进程装配：`cmd/server/main.go`
- 路由：`internal/httpapi/router.go`
- 响应：`internal/httpapi/result.go` 的 `WriteOK`
- 请求体上限：`internal/httpapi/body_limit.go`

`http.Server` 设置了 `ReadHeaderTimeout` 5 秒、`ReadTimeout` 60 秒、`WriteTimeout` 60 秒、`IdleTimeout` 120 秒。路由最外层用 `http.MaxBytesReader`，默认 8MiB。超过上限时解码失败，接口返回 400「请求体无效」。

成功响应是 `{code:0, msg:"success", data}`。Java `CommonResult.success` 的 `msg` 是 `success`。错误仍走 `apperr`，HTTP 状态和 `code` 分开。

## 和 Java 的对应

Java 用 Spring MVC 的容器超时和 `CommonResult`。Go 没有照搬 Tomcat 的全部配置，只把缺超时、缺请求体上限、成功文案写成 `ok` 这三处补上。

## 不要这样做

不要在 handler 里再包一层不同的成功文案。前端和测试都认 `msg=success`。
