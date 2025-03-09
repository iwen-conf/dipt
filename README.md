# 🐳 DIPT (Docker Image Pull Tar)

一个无需 Docker 环境即可拉取 Docker 镜像并保存为 tar 文件的 Go 工具。

## ✨ 功能特点

- 🚀 无需安装 Docker 即可拉取镜像
- 🔄 支持从公共或私有 Docker Registry 拉取镜像
- 💾 将镜像保存为标准 tar 文件，可用于离线环境
- �� 支持认证信息配置，可访问私有仓库
- 🎯 支持指定目标操作系统和架构
- 📝 智能文件命名，自动包含镜像信息
- 🛠️ 轻量级命令行工具，易于使用

## 📥 安装

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/iwen-conf/dipt.git
cd dipt

# 编译
go build -o dipt

# 可选：将编译好的二进制文件移动到PATH路径
mv dipt /usr/local/bin/
```

## 📚 使用方法

### 基本用法

```bash
# 基本命令格式
dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]

# 使用当前系统的平台拉取镜像
dipt nginx:latest
# 将生成类似 nginx_latest_darwin_arm64.tar 的文件（基于当前系统）

# 指定目标平台拉取镜像
dipt -os linux -arch amd64 nginx:latest
# 将生成 nginx_latest_linux_amd64.tar

# 指定目标平台和自定义输出文件名
dipt -os linux -arch arm64 nginx:latest custom.tar
```

### 文件命名规则

如果不指定输出文件名，程序会自动生成格式为 `软件名_版本_系统_架构.tar` 的文件名，例如：
- `nginx_latest_linux_amd64.tar`
- `mysql_8.0_linux_arm64.tar`
- `ubuntu_22.04_linux_amd64.tar`

### 支持的平台

可以通过 `-os` 和 `-arch` 参数指定目标平台：

- 操作系统 (-os)：
  - linux
  - windows
  - darwin (macOS)
  
- 架构 (-arch)：
  - amd64 (x86_64)
  - arm64 (aarch64)
  - arm
  - 386 (x86)

### 使用私有仓库

创建 `config.json` 文件，包含私有仓库的认证信息：

```json
{
  "registry": {
    "username": "your-username",
    "password": "your-password"
  }
}
```

然后拉取私有仓库中的镜像：

```bash
dipt -os linux -arch amd64 private-registry.example.com/myapp:1.0
# 将生成 myapp_1.0_linux_amd64.tar
```

## ⚙️ 配置文件

配置文件 `config.json` 应放在与程序相同的目录下，格式如下：

```json
{
  "registry": {
    "username": "your-username",
    "password": "your-password"
  }
}
```

如果不提供配置文件或配置文件中没有认证信息，程序将使用匿名访问拉取公共镜像。

## 🌟 优势

- **🔥 无需 Docker 环境**：在无法安装 Docker 的环境中也能拉取镜像
- **🎯 跨平台支持**：可以在任何平台上拉取任意平台的镜像
- **📦 离线部署支持**：可以在有网络的环境中拉取镜像，然后将 tar 文件传输到离线环境使用
- **🏷️ 智能命名**：自动生成包含完整信息的文件名
- **⚡ 轻量级**：只依赖 Go 标准库和容器注册表交互库
- **🔒 安全**：不需要 Docker daemon 权限

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 👥 贡献

欢迎提交问题和贡献代码！

- GitHub: [https://github.com/iwen-conf/dipt.git](https://github.com/iwen-conf/dipt.git)
- 联系邮箱: iluwenconf@163.com

让我们一起改进这个工具！🚀
