package docker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"dipt/internal/errors"
	"dipt/internal/retry"
	"dipt/internal/types"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	v1types "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// PullOptions 拉取选项
type PullOptions struct {
	ImageName  string
	OutputFile string
	Platform   types.Platform
	Config     types.Config
	OnProgress ProgressCallback          // 进度回调
	OnLog      func(level, msg string)   // 日志回调
}

// logMsg 发送日志消息
func (o *PullOptions) logMsg(level, format string, args ...interface{}) {
	if o.OnLog != nil {
		o.OnLog(level, fmt.Sprintf(format, args...))
	}
}

// PullAndSave 拉取镜像并保存为 tar 文件（新接口）
func PullAndSave(opts PullOptions) error {
	// 检查是否为演练模式
	if os.Getenv("DIPT_DRY_RUN") == "1" {
		opts.logMsg("info", "[演练模式] 将拉取镜像 %s 并保存到 %s", opts.ImageName, opts.OutputFile)
		opts.logMsg("info", "[演练模式] 平台: %s/%s", opts.Platform.OS, opts.Platform.Arch)
		opts.logMsg("success", "[演练模式] 检测完成，未执行实际操作")
		return nil
	}

	ref, err := name.ParseReference(opts.ImageName)
	if err != nil {
		return errors.NewImageNotFoundError(opts.ImageName, err)
	}

	var auth authn.Authenticator
	if opts.Config.Registry.Username != "" && opts.Config.Registry.Password != "" {
		auth = authn.FromConfig(authn.AuthConfig{
			Username: opts.Config.Registry.Username,
			Password: opts.Config.Registry.Password,
		})
	} else {
		auth = authn.Anonymous
	}

	options := []remote.Option{remote.WithAuth(auth)}

	// 读取超时配置
	timeout := 120 * time.Second
	if t := os.Getenv("DIPT_TIMEOUT"); t != "" {
		if sec, perr := strconv.Atoi(t); perr == nil && sec > 0 {
			timeout = time.Duration(sec) * time.Second
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	options = append(options, remote.WithContext(ctx))

	// 添加平台选项
	options = append(options, remote.WithPlatform(v1.Platform{
		OS:           opts.Platform.OS,
		Architecture: opts.Platform.Arch,
	}))

	// 处理自定义镜像源
	config := opts.Config
	if customMirror := os.Getenv("DIPT_CUSTOM_MIRROR"); customMirror != "" {
		config.Registry.Mirrors = append([]string{customMirror}, config.Registry.Mirrors...)
		opts.logMsg("info", "使用自定义镜像源: %s", customMirror)
	}

	// 尝试使用镜像加速器
	if len(config.Registry.Mirrors) > 0 && IsDockerHubImage(ref) {
		mirrorManager := NewMirrorManager(config.Registry.Mirrors)

		retryConfig := retry.DefaultConfig()
		retryConfig.MaxRetries = 2

		// mirror 日志回调，将 mirror 状态信息传递到 TUI
		mirrorLogFunc := func(level, msg string) {
			opts.logMsg(level, "%s", msg)
		}

		// mirror 使用匿名认证（不传原始 registry 的凭据）
		mirrorOptions := make([]remote.Option, 0, len(options))
		for _, opt := range options {
			mirrorOptions = append(mirrorOptions, opt)
		}
		// 替换 auth 为匿名认证
		mirrorOptions = append(mirrorOptions, remote.WithAuth(authn.Anonymous))

		err = mirrorManager.TryPullWithMirrors(ref, mirrorOptions, mirrorLogFunc, func(mirrorRef name.Reference, mirrorURL string) error {
			return retry.WithRetry(func() error {
				desc, err := remote.Get(mirrorRef, mirrorOptions...)
				if err != nil {
					return err
				}
				return downloadAndSave(mirrorRef, opts.OutputFile, desc, mirrorOptions, &opts)
			}, retryConfig, fmt.Sprintf("拉取镜像 [%s]", mirrorURL))
		})

		if err == nil {
			return nil
		}
		opts.logMsg("warning", "镜像加速器失败，尝试使用原始地址")
	}

	// 使用原始地址
	retryConfig := retry.DefaultConfig()
	var desc *remote.Descriptor
	err = retry.WithRetry(func() error {
		var getErr error
		desc, getErr = remote.Get(ref, options...)
		return getErr
	}, retryConfig, fmt.Sprintf("获取镜像元数据 [%s]", opts.ImageName))

	if err != nil {
		// 特殊处理 docker.dragonflydb.io
		if strings.Contains(err.Error(), "ghcr.io") && strings.Contains(opts.ImageName, "docker.dragonflydb.io") {
			opts.logMsg("info", "检测到 docker.dragonflydb.io 重定向到 ghcr.io，尝试直接使用 ghcr.io 地址...")
			newImageName := strings.Replace(opts.ImageName, "docker.dragonflydb.io", "ghcr.io", 1)
			newRef, parseErr := name.ParseReference(newImageName)
			if parseErr == nil {
				err = retry.WithRetry(func() error {
					var getErr error
					desc, getErr = remote.Get(newRef, options...)
					return getErr
				}, retryConfig, "获取镜像元数据 [ghcr.io]")

				if err == nil {
					ref = newRef
					opts.logMsg("success", "成功使用 ghcr.io 地址: %s", newImageName)
					return downloadAndSave(ref, opts.OutputFile, desc, options, &opts)
				}
			}
		}

		// GitHub Container Registry 认证问题
		registry := ref.Context().Registry.Name()
		if (registry == "ghcr.io" || strings.Contains(opts.ImageName, "docker.dragonflydb.io")) && strings.Contains(err.Error(), "DENIED") {
			return &errors.DiptError{
				Type: errors.ErrorUnauthorized,
				Message: fmt.Sprintf("访问 GitHub Container Registry 需要认证\n建议：\n"+
					"1. 对于 DragonflyDB，请使用正确的镜像地址: ghcr.io/dragonflydb/dragonfly:latest\n"+
					"2. GitHub Container Registry 可能需要 GitHub Personal Access Token\n"+
					"3. 在 config.json 中配置 GitHub 用户名和 Personal Access Token\n"+
					"原始错误: %v", err),
				Err: err,
			}
		}

		if errors.IsManifestUnknownError(err) {
			return errors.NewPlatformNotSupportedError(opts.ImageName, opts.Platform.OS, opts.Platform.Arch, err)
		} else if errors.IsUnauthorizedError(err) {
			return errors.NewUnauthorizedError(ref.Context().RegistryStr(), err)
		} else if errors.IsNetworkError(err) {
			return errors.NewNetworkError(err)
		}
		return errors.NewImageNotFoundError(opts.ImageName, err)
	}

	return downloadAndSave(ref, opts.OutputFile, desc, options, &opts)
}

// downloadAndSave 下载并保存镜像
func downloadAndSave(ref name.Reference, outputFile string, desc *remote.Descriptor, options []remote.Option, opts *PullOptions) error {
	metaImg, err := desc.Image()
	if err != nil {
		if errors.IsPlatformNotSupportedError(err) {
			return errors.NewPlatformNotSupportedError(ref.Name(), "", "", err)
		}
		return fmt.Errorf("获取镜像元数据失败: %v", err)
	}
	m, err := metaImg.Manifest()
	if err != nil {
		return fmt.Errorf("获取镜像清单失败: %v", err)
	}

	var totalSize int64
	totalSize += m.Config.Size
	for _, l := range m.Layers {
		if !v1types.MediaType(l.MediaType).IsDistributable() {
			continue
		}
		totalSize += l.Size
	}

	opts.logMsg("info", "镜像总大小: %s", formatBytesStr(totalSize))

	// 使用带总量追踪的 RoundTripper
	rt := NewTotalTrackingRoundTripper(http.DefaultTransport, totalSize, opts.OnProgress)
	dlOptions := append(options, remote.WithTransport(rt))

	img, err := remote.Image(ref, dlOptions...)
	if err != nil {
		if errors.IsPlatformNotSupportedError(err) {
			return errors.NewPlatformNotSupportedError(ref.Name(), "", "", err)
		} else if errors.IsUnauthorizedError(err) {
			return errors.NewUnauthorizedError(ref.Context().RegistryStr(), err)
		} else if errors.IsNetworkError(err) {
			return errors.NewNetworkError(err)
		}
		return fmt.Errorf("拉取镜像失败: %v", err)
	}

	err = tarball.WriteToFile(outputFile, ref, img)
	if err != nil {
		return fmt.Errorf("保存镜像到 tar 文件失败: %v", err)
	}

	// 报告 100% 进度
	if opts.OnProgress != nil {
		opts.OnProgress(totalSize, totalSize)
	}
	opts.logMsg("success", "镜像已保存到 %s", outputFile)
	return nil
}

func formatBytesStr(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ParseImageName 从镜像名称中提取软件名和版本
func ParseImageName(imageName string) (software, version string) {
	parts := strings.Split(imageName, ":")
	if len(parts) < 2 {
		return CleanSoftwareName(parts[0]), "latest"
	}
	software = CleanSoftwareName(parts[0])
	return software, parts[1]
}

// CleanSoftwareName 处理软件名称，替换斜杠为下划线
func CleanSoftwareName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

// GenerateOutputFileName 生成输出文件名
func GenerateOutputFileName(imageName string, platform types.Platform) string {
	software, version := ParseImageName(imageName)
	return fmt.Sprintf("%s_%s_%s_%s.tar", software, version, platform.OS, platform.Arch)
}
