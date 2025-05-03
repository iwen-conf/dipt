package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/schollz/progressbar/v3"
)

// pullAndSaveImage 拉取镜像并保存为 tar 文件，带进度显示
func pullAndSaveImage(imageName, outputFile string, platform Platform, config Config) error {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return NewImageNotFoundError(imageName, err)
	}

	var auth authn.Authenticator
	if config.Registry.Username != "" && config.Registry.Password != "" {
		auth = authn.FromConfig(authn.AuthConfig{
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		})
	} else {
		auth = authn.Anonymous
	}

	options := []remote.Option{remote.WithAuth(auth)}

	// 添加平台选项
	options = append(options, remote.WithPlatform(v1.Platform{
		OS:           platform.OS,
		Architecture: platform.Arch,
	}))

	// 尝试使用配置的镜像加速器
	var lastErr error
	if len(config.Registry.Mirrors) > 0 && isDockerHubImage(ref) {
		for _, mirror := range config.Registry.Mirrors {
			mirrorRef, err := createMirrorRef(ref, mirror)
			if err != nil {
				continue
			}

			desc, err := remote.Get(mirrorRef, options...)
			if err == nil {
				ref = mirrorRef
				fmt.Printf("使用镜像加速器: %s\n", mirror)
				return downloadAndSaveImage(ref, outputFile, desc, options)
			}
			lastErr = err
		}
	}

	// 如果镜像加速器都失败了，或者不是 Docker Hub 镜像，使用原始地址
	desc, err := remote.Get(ref, options...)
	if err != nil {
		if lastErr != nil {
			fmt.Printf("镜像加速器访问失败，尝试使用原始地址\n")
		}
		if IsManifestUnknownError(err) {
			return NewPlatformNotSupportedError(imageName, platform.OS, platform.Arch, err)
		} else if IsUnauthorizedError(err) {
			return NewUnauthorizedError(ref.Context().RegistryStr(), err)
		} else if IsNetworkError(err) {
			return NewNetworkError(err)
		}
		return NewImageNotFoundError(imageName, err)
	}

	return downloadAndSaveImage(ref, outputFile, desc, options)
}

// downloadAndSaveImage 下载并保存镜像
func downloadAndSaveImage(ref name.Reference, outputFile string, desc *remote.Descriptor, options []remote.Option) error {
	img, err := desc.Image()
	if err != nil {
		if IsPlatformNotSupportedError(err) {
			return NewPlatformNotSupportedError(ref.Name(), "", "", err)
		}
		return fmt.Errorf("获取镜像失败: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("获取层信息失败: %v", err)
	}

	var totalSize int64
	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return fmt.Errorf("获取层大小失败: %v", err)
		}
		totalSize += size
	}

	bar := progressbar.NewOptions64(totalSize,
		progressbar.OptionSetDescription("拉取镜像中"),
		progressbar.OptionShowBytes(true),
	)

	rt := &progressRoundTripper{rt: http.DefaultTransport, bar: bar}
	options = append(options, remote.WithTransport(rt))
	img, err = remote.Image(ref, options...)
	if err != nil {
		if IsPlatformNotSupportedError(err) {
			return NewPlatformNotSupportedError(ref.Name(), "", "", err)
		} else if IsUnauthorizedError(err) {
			return NewUnauthorizedError(ref.Context().RegistryStr(), err)
		} else if IsNetworkError(err) {
			return NewNetworkError(err)
		}
		return fmt.Errorf("拉取镜像失败: %v", err)
	}

	err = tarball.WriteToFile(outputFile, ref, img)
	if err != nil {
		return fmt.Errorf("保存镜像到 tar 文件失败: %v", err)
	}

	return nil
}

// isDockerHubImage 判断是否是 Docker Hub 镜像
func isDockerHubImage(ref name.Reference) bool {
	registry := ref.Context().Registry.Name()
	return registry == "docker.io" || registry == "registry-1.docker.io"
}

// createMirrorRef 创建镜像加速器引用
func createMirrorRef(ref name.Reference, mirror string) (name.Reference, error) {
	// 移除协议前缀
	mirror = strings.TrimPrefix(mirror, "http://")
	mirror = strings.TrimPrefix(mirror, "https://")
	mirror = strings.TrimSuffix(mirror, "/")

	// 获取原始镜像名称和标签
	originalName := ref.Context().RepositoryStr()
	originalName = strings.TrimPrefix(originalName, "library/")
	tag := ref.Identifier()

	// 构建新的引用
	newRef := fmt.Sprintf("%s/%s:%s", mirror, originalName, tag)
	return name.ParseReference(newRef)
}

// 从镜像名称中提取软件名和版本
func parseImageName(imageName string) (software, version string) {
	parts := strings.Split(imageName, ":")
	if len(parts) < 2 {
		return cleanSoftwareName(parts[0]), "latest"
	}

	// 处理软件名称中可能包含的路径
	software = cleanSoftwareName(parts[0])

	return software, parts[1]
}

// cleanSoftwareName 处理软件名称，替换斜杠为下划线，避免被当作目录分隔符
func cleanSoftwareName(name string) string {
	// 替换所有斜杠为下划线
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

// generateOutputFileName 生成输出文件名
func generateOutputFileName(imageName string, platform Platform) string {
	software, version := parseImageName(imageName)
	return fmt.Sprintf("%s_%s_%s_%s.tar", software, version, platform.OS, platform.Arch)
}
