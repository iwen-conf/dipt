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

	desc, err := remote.Get(ref, options...)
	if err != nil {
		if IsManifestUnknownError(err) {
			// 检查是否是平台不支持的错误
			return NewPlatformNotSupportedError(imageName, platform.OS, platform.Arch, err)
		} else if IsUnauthorizedError(err) {
			// 检查是否是认证错误
			return NewUnauthorizedError(ref.Context().RegistryStr(), err)
		} else if IsNetworkError(err) {
			// 检查是否是网络错误
			return NewNetworkError(err)
		}
		return NewImageNotFoundError(imageName, err)
	}

	img, err := desc.Image()
	if err != nil {
		if IsPlatformNotSupportedError(err) {
			// 检查是否是平台不支持的错误
			return NewPlatformNotSupportedError(imageName, platform.OS, platform.Arch, err)
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
			return NewPlatformNotSupportedError(imageName, platform.OS, platform.Arch, err)
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

// 从镜像名称中提取软件名和版本
func parseImageName(imageName string) (software, version string) {
	parts := strings.Split(imageName, ":")
	if len(parts) < 2 {
		return parts[0], "latest"
	}

	// 处理软件名称中可能包含的路径
	nameParts := strings.Split(parts[0], "/")
	software = nameParts[len(nameParts)-1]

	return software, parts[1]
}

// generateOutputFileName 生成输出文件名
func generateOutputFileName(imageName string, platform Platform) string {
	software, version := parseImageName(imageName)
	return fmt.Sprintf("%s_%s_%s_%s.tar", software, version, platform.OS, platform.Arch)
}
