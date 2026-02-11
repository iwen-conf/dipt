<div align="center">

# dipt

**Docker 镜像拉取工具**

无需 Docker daemon，直接拉取镜像并保存为 tar 文件。

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

[English](README.md) · [中文](README_zh.md)

</div>

---

## 为什么用 dipt？

有时你需要 Docker 镜像，但无法运行 `docker pull` — 没有 daemon、网络受限、离线环境。**dipt** 通过 [go-containerregistry](https://github.com/google/go-containerregistry) 直接与容器仓库通信，将镜像保存为标准 tar 文件，随时 `docker load` 导入。

## 特性

- **交互式 TUI** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea)，实时进度条与日志查看
- **镜像加速器** — 自动探测、健康检查、逐个回退
- **多平台** — linux / windows / darwin × amd64 / arm64 / arm / 386
- **私有仓库** — 支持用户名/密码认证，密码安全输入
- **智能重试** — 指数退避 + 随机抖动，应对瞬时故障
- **三层配置** — 环境变量 > 项目配置 > 用户配置

## 快速开始

```bash
# 构建
git clone <repo-url> && cd dipt
go build -o dipt ./cmd/dipt

# 启动 TUI
./dipt
```

首次运行会进入配置向导。之后主菜单提供三个入口：

| 菜单 | 功能 |
|------|------|
| **拉取镜像** | 输入镜像名，选择平台，下载为 `.tar` |
| **设置** | 默认 OS、架构、保存目录、仓库凭据 |
| **镜像源管理** | 添加、删除、测试镜像加速器 |

## 快捷键

| 按键 | 操作 |
|------|------|
| `↑↓` / `jk` | 导航 |
| `Enter` | 确认 |
| `Tab` / `Shift+Tab` | 下一个 / 上一个字段 |
| `←→` | 切换选项 |
| `Esc` | 返回 |
| `q` / `Ctrl+C` | 退出 |

## 配置

### 用户配置 `~/.dipt_config`

首次运行时由配置向导自动创建：

```json
{
  "default_os": "linux",
  "default_arch": "amd64",
  "default_save_dir": "./images",
  "registry": {
    "mirrors": ["https://mirror.example.com"],
    "username": "",
    "password": ""
  }
}
```

### 项目配置 `./config.json`

可选的项目级覆盖，优先级高于用户配置，结构相同。

### 环境变量

| 变量 | 说明 |
|------|------|
| `DIPT_DEFAULT_OS` | 默认操作系统 |
| `DIPT_DEFAULT_ARCH` | 默认架构 |
| `DIPT_DEFAULT_SAVE_DIR` | 默认保存目录 |
| `DIPT_REGISTRY_MIRRORS` | 镜像源 URL（逗号分隔） |
| `DIPT_REGISTRY_USERNAME` | 仓库用户名 |
| `DIPT_REGISTRY_PASSWORD` | 仓库密码 |
| `DIPT_CUSTOM_MIRROR` | 自定义镜像源（优先使用） |
| `DIPT_TIMEOUT` | 超时秒数（默认 `120`） |
| `DIPT_NO_INTERACTIVE=1` | 跳过配置向导 |
| `DIPT_DRY_RUN=1` | 演练模式 |

> 优先级：环境变量 > `./config.json` > `~/.dipt_config`

## 许可证

[MIT](LICENSE)
