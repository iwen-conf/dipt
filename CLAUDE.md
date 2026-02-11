[根目录](../../CLAUDE.md) > **dipt**

# dipt

## 模块职责

Docker 镜像拉取与保存 CLI 工具。在无法直接使用 `docker pull` 的环境中（如无 Docker daemon、网络受限），通过 go-containerregistry 直接从 Registry 拉取镜像并保存为 tar 文件。支持镜像加速器自动探测与回退、多平台选择、交互式首次配置。

## 入口与启动

- 入口: `cmd/dipt/main.go`
- 流程: 加载配置 -> 解析 CLI 参数 -> 初始化日志 -> 拉取镜像 -> 保存 tar -> 打印总结

## 对外接口（CLI）

```
dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]
dipt set <os|arch|save_dir> <值>
dipt -conf new
dipt mirror list|add|del|clear|test <URL>
```

选项: `--verbose`, `--mirror=<URL>`, `--fix`, `--dry-run`

环境变量:
- `DIPT_NO_INTERACTIVE=1` - 禁用交互式配置
- `DIPT_DEFAULT_OS` / `DIPT_DEFAULT_ARCH` / `DIPT_DEFAULT_SAVE_DIR`
- `DIPT_REGISTRY_USERNAME` / `DIPT_REGISTRY_PASSWORD` / `DIPT_REGISTRY_MIRRORS`
- `DIPT_CUSTOM_MIRROR` - 指定自定义镜像源
- `DIPT_TIMEOUT` - 超时秒数（默认 120）
- `DIPT_DRY_RUN=1` - 演练模式

## 关键依赖与配置

核心依赖:
- `github.com/google/go-containerregistry` - Registry 交互、镜像拉取与 tarball 写入
- `github.com/schollz/progressbar/v3` - 终端进度条
- `golang.org/x/term` - 终端检测（判断是否交互式环境）

配置层级（优先级从高到低）:
1. 环境变量（`DIPT_*`）
2. 项目配置（`./config.json`）
3. 用户配置（`~/.dipt_config`）

## 数据模型

定义在 `internal/types/types.go`:
- `Config` - 项目级配置，包含 `Registry`（mirrors/username/password）
- `UserConfig` - 用户级配置，包含默认 OS/Arch/SaveDir 及 Registry
- `Platform` - 平台信息（OS + Arch）

## 目录结构

```
dipt/
  cmd/dipt/main.go          # CLI 入口
  internal/
    cli/cli.go               # 参数解析、子命令路由、配置模板生成
    color/color.go            # ANSI 彩色终端输出工具
    color/color_test.go       # color 包单元测试
    config/config.go          # 配置加载/保存/合并、镜像源管理命令
    config/interactive.go     # 首次运行交互式配置向导
    errors/errors.go          # 自定义错误类型与错误分类判断
    logger/logger.go          # 单例日志器，彩色输出与统计
    retry/retry.go            # 指数退避重试机制
    retry/retry_test.go       # retry 包单元测试
    types/types.go            # 共享数据结构定义
  pkg/docker/
    image.go                  # 核心: 镜像拉取、保存、镜像源回退
    mirror.go                 # 镜像源管理器: 探测、排序、回退拉取
    progress.go               # HTTP RoundTripper 进度条包装
  go.mod
  go.sum
  LICENSE                     # MIT
  .gitignore
```

## 测试与质量

- `internal/color/color_test.go` - 测试颜色输出与 NO_COLOR 环境变量支持
- `internal/retry/retry_test.go` - 测试重试机制、退避计算、RetryableOperation 接口
- 运行: `cd dipt && go test ./...`
- 缺口: `pkg/docker/`、`internal/cli/`、`internal/config/`、`internal/errors/` 无测试

## 常见问题 (FAQ)

- 首次运行会触发交互式配置（可通过 `DIPT_NO_INTERACTIVE=1` 跳过）
- 镜像加速器仅对 Docker Hub 镜像生效（`docker.io` / `registry-1.docker.io`）
- 拉取失败时会自动重试（默认 3 次，指数退避），镜像源逐个回退后尝试原始地址
- 支持 `docker.dragonflydb.io` 到 `ghcr.io` 的自动重定向处理

## 相关文件清单

| 文件 | 职责 |
|------|------|
| `cmd/dipt/main.go` | 程序入口 |
| `internal/cli/cli.go` | CLI 参数解析与子命令 |
| `internal/config/config.go` | 配置管理核心 |
| `internal/config/interactive.go` | 交互式配置向导 |
| `internal/types/types.go` | 数据结构定义 |
| `internal/errors/errors.go` | 错误类型与分类 |
| `internal/logger/logger.go` | 日志系统 |
| `internal/retry/retry.go` | 重试机制 |
| `internal/color/color.go` | 彩色输出 |
| `pkg/docker/image.go` | 镜像拉取与保存 |
| `pkg/docker/mirror.go` | 镜像源管理 |
| `pkg/docker/progress.go` | 下载进度条 |

## 变更记录 (Changelog)

| 时间 | 操作 |
|------|------|
| 2026-02-11T18:07:23 | 初始扫描生成 |
