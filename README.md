# 🐳 DIPT (Docker Image Pull Tar)

一个无需 Docker 环境即可拉取 Docker 镜像并保存为 tar 文件的 Go 工具。

## ✨ 功能特点

- 🚀 无需安装 Docker 即可拉取镜像
- 🔄 支持从公共或私有 Docker Registry 拉取镜像
- 💾 将镜像保存为标准 tar 文件，可用于离线环境
- 🔐 支持认证信息配置，可访问私有仓库
- 🎯 支持指定目标操作系统和架构
- 📝 智能文件命名，自动包含镜像信息
- 🛠️ 轻量级命令行工具，易于使用
- ⚠️ 友好的错误提示，帮助快速定位问题
- ⚙️ 支持交互式配置和默认值设置
- 🚄 支持镜像加速器配置，提升下载速度

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

### 镜像加速器管理

DIPT 提供了命令行工具来管理镜像加速器，无需手动编辑配置文件：

```bash
# 列出所有配置的镜像加速器
dipt mirror list

# 添加新的镜像加速器
dipt mirror add https://mirror.example.com

# 删除指定的镜像加速器
dipt mirror del https://mirror.example.com

# 清空所有镜像加速器
dipt mirror clear
```

默认提供以下镜像加速器：
- https://registry.docker-cn.com
- https://docker.mirrors.ustc.edu.cn
- http://hub-mirror.c.163.com

> 注意：镜像加速器仅对 Docker Hub 镜像有效，私有仓库镜像不会使用加速器。

### 配置文件管理

DIPT 支持两种配置文件：

1. 全局配置文件：`~/.dipt_config`
   - 存储默认设置和全局镜像加速器配置
   - 由 `dipt set` 和 `dipt mirror` 命令管理

2. 项目配置文件：`./config.json`
   - 存储项目特定的配置，如私有仓库认证信息
   - 可以使用 `dipt -conf new` 生成模板

配置文件优先级：
- 优先使用当前目录下的 `config.json`
- 如果当前目录没有配置文件，则使用 `~/.dipt_config`

生成配置模板：
```bash
# 在当前目录生成配置文件模板
dipt -conf new
```

示例配置文件格式：
```json
{
  "registry": {
    "mirrors": [
      "https://registry.docker-cn.com",
      "https://docker.mirrors.ustc.edu.cn",
      "http://hub-mirror.c.163.com"
    ],
    "username": "your-username",
    "password": "your-password"
  }
}
```

### 文件命名规则

如果不指定输出文件名，程序会自动生成格式为 `软件名_版本_系统_架构.tar` 的文件名，例如：
- `nginx_latest_linux_amd64.tar`
- `mysql_8.0_linux_arm64.tar`
- `ubuntu_22.04_linux_amd64.tar`

### 支持的平台

可以通过 `-os` 和 `-arch` 参数指定目标平台：

- 操作系统 (-os)：
  - linux（最广泛支持）
  - windows（部分镜像支持）
  - darwin（部分镜像支持）
  
- 架构 (-arch)：
  - amd64 (x86_64)（最广泛支持）
  - arm64 (aarch64)
  - arm
  - 386 (x86)

> ⚠️ **注意**：并非所有镜像都支持所有平台组合。例如：
> - `mysql:latest` 不支持 windows 平台
> - 某些镜像可能只支持特定的架构
> - 建议优先使用 linux 平台的镜像，它们通常有最好的兼容性

### 错误处理

当遇到错误时，程序会提供清晰的错误信息。以下是一些常见错误及解决方案：

1. 平台不支持
```bash
$ dipt -os windows -arch amd64 mysql:latest
错误: 获取镜像描述失败: GET https://registry-1.docker.io/v2/library/mysql/manifests/latest: MANIFEST_UNKNOWN: manifest unknown
```
解决方案：
- 检查镜像是否支持指定的平台
- 使用 `linux` 平台替代
- 尝试其他版本的镜像

2. 镜像不存在
```bash
$ dipt nginx:nonexist
错误: 获取镜像描述失败: GET https://registry-1.docker.io/v2/library/nginx/manifests/nonexist: MANIFEST_UNKNOWN: manifest unknown
```
解决方案：
- 检查镜像名称和版本是否正确
- 在 Docker Hub 或相应的镜像仓库中验证镜像是否存在

3. 私有仓库认证失败
```bash
$ dipt private-registry.example.com/myapp:1.0
错误: 获取镜像描述失败: GET https://private-registry.example.com/v2/myapp/manifests/1.0: UNAUTHORIZED: authentication required
```
解决方案：
- 检查 `config.json` 文件是否存在且格式正确
- 验证用户名和密码是否正确
- 确认是否有权限访问该镜像

4. 网络问题
```bash
$ dipt nginx:latest
错误: 获取镜像描述失败: Get https://registry-1.docker.io/v2/: dial tcp: lookup registry-1.docker.io: no such host
```
解决方案：
- 检查网络连接
- 验证是否需要配置代理
- 确认 DNS 解析是否正常

### 最佳实践

1. **平台选择**：
   - 优先使用 `linux` 平台的镜像
   - 对于 Windows 应用，确认镜像是否有 Windows 版本
   - 选择与目标环境匹配的架构

2. **版本选择**：
   - 建议使用具体的版本号而不是 `latest` 标签
   - 在生产环境中使用固定版本以确保稳定性

3. **镜像加速**：
   - 对于 Docker Hub 镜像，建议配置镜像加速器
   - 选择延迟最低的镜像加速器
   - 可以配置多个加速器作为备用

4. **错误处理**：
   - 遇到平台不支持错误时，先检查 Docker Hub 上的支持信息
   - 如果镜像拉取速度慢，尝试使用镜像加速器
   - 使用 `-os` 和 `-arch` 参数时要谨慎，确保目标平台受支持

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

## ⚙️ 配置管理

### 首次运行配置

首次运行 DIPT 时，程序会自动启动交互式配置向导，引导您设置：

1. 默认操作系统（linux/windows/darwin）
2. 默认架构（amd64/arm64/arm/386）
3. 默认保存目录

配置示例：
```bash
👋 欢迎使用 DIPT！
📝 首次运行需要进行一些基本设置...

💻 请选择默认的操作系统 [linux/windows/darwin] (默认: linux): linux
🔧 请选择默认的架构 [amd64/arm64/arm/386] (默认: amd64): arm64
📂 请输入默认的镜像保存目录 (默认: /Users/username/DockerImages): 

✅ 配置完成！配置文件已保存到: .dipt_config
```

### 修改默认配置

您可以使用 `dipt set` 命令随时修改默认配置：

```bash
# 修改默认操作系统
dipt set os linux

# 修改默认架构
dipt set arch arm64

# 修改默认保存目录
dipt set save_dir ~/docker-images
```

### 配置文件位置

用户配置文件保存在 `~/.dipt_config`，使用 JSON 格式：

```json
{
  "default_os": "linux",
  "default_arch": "arm64",
  "default_save_dir": "/Users/username/DockerImages"
}
```

### 使用默认配置

设置了默认配置后，您可以省略 `-os` 和 `-arch` 参数，DIPT 将使用配置文件中的默认值：

```bash
# 使用默认配置拉取镜像
dipt nginx:latest
# 将使用配置的默认操作系统和架构

# 覆盖默认配置
dipt -os windows -arch amd64 nginx:latest
# 临时使用指定的平台，不会修改默认配置
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
- **❗ 错误处理**：提供清晰的错误信息和解决方案

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 👥 贡献

欢迎提交问题和贡献代码！

- GitHub: [https://github.com/iwen-conf/dipt.git](https://github.com/iwen-conf/dipt.git)
- 联系邮箱: iluwenconf@163.com

让我们一起改进这个工具！🚀
