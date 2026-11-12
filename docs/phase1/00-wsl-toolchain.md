# 00 · WSL 工具链（一次装好）

本学习仓在 **WSL2** 里编译运行 Go。Windows 侧负责 Docker Desktop + 读 Java 对照仓。

## 1. 进入仓库

```bash
cd /mnt/e/workspace/GoProject/ai-fusion-video-study
```

若磁盘挂载权限别扭，可把仓 clone 到 `~/code/ai-fusion-video-go`（GitHub remote：`zhazhahehedian/ai-fusion-video-go`）。

## 2. 安装 Go 1.23+

```bash
# 示例：官方 tarball（版本号按 go.dev/dl 最新 1.23.x / 1.24.x 调整）
curl -fsSL https://go.dev/dl/go1.23.6.linux-amd64.tar.gz -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

## 3. 安装 sqlc / goose

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
sqlc version
goose -version
```

## 4. 拉取依赖并跑空壳

```bash
cd /mnt/e/workspace/GoProject/ai-fusion-video-study
go mod tidy
make run
# 另开终端：
curl -s http://127.0.0.1:18080/healthz
# 期望: {"code":0,"msg":"ok","data":{"status":"up","phase":1,...}}
```

## 5. 推到 GitHub

```bash
git remote -v
# 应为 https://github.com/zhazhahehedian/ai-fusion-video-go.git
git push -u origin main
```

未登录时用 `gh auth login`，或在 Cursor 里连接 GitHub。

## 勾选

- [ ] `go version` ≥ 1.23
- [ ] `sqlc` / `goose` 在 PATH
- [ ] `curl localhost:18080/healthz` 返回 `code:0`
- [ ] `git push -u origin main` 成功
