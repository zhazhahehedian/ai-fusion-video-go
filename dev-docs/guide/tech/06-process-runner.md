# 外部进程

## 要解决什么

媒体探测会启动 ffmpeg 或其他命令。取消和租约失效时要杀掉它。命令若不停输出，不能让内存一直涨。杀掉之后如果进程还不退出，不能和仍在写的缓冲区并发读。

## 在哪里

`internal/media/process.go` 的 `ProcessRunner`。

- `MaxOutputBytes` 默认 1MiB。超出的字节被丢掉，管道继续被读取，进程不会因为管道写满而卡住。
- `Write` 和 `String` 共用一把锁。
- `killWait` 先 `Kill`，再等 `Wait` 最多 5 秒。超时返回「进程在终止后仍未退出」，这次调用不再读取输出。

`onTick` 用来在命令运行期间续租或发现租约丢失。返回错误时走同一条终止路径。

## 和 Java 的对应

Java `MediaProcessRunner` 同样在超时或取消时销毁进程，并限制收集的输出。Go 这里没有接合成视频、对象存储或真实 ffmpeg 滤镜链。`FFMPEG_VERSION` 只用来探测进程能否被取消。
