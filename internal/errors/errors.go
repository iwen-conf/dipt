package errors

import (
    "errors"
    "fmt"
    "strings"

    "github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ErrorType 定义错误类型
type ErrorType int

const (
	ErrorUnknown ErrorType = iota
	ErrorPlatformNotSupported
	ErrorImageNotFound
	ErrorUnauthorized
	ErrorNetwork
)

// DiptError 自定义错误类型
type DiptError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *DiptError) Error() string {
	return e.Message
}

// NewPlatformNotSupportedError 创建平台不支持错误
func NewPlatformNotSupportedError(imageName, os, arch string, err error) *DiptError {
	return &DiptError{
		Type: ErrorPlatformNotSupported,
		Message: fmt.Sprintf("镜像 %s 不支持平台 %s/%s\n建议：\n"+
			"1. 检查镜像是否支持该平台组合\n"+
			"2. 尝试使用 linux 平台（最广泛支持）\n"+
			"3. 访问 https://hub.docker.com 查看镜像支持的平台\n"+
			"4. 尝试其他版本的镜像\n"+
			"原始错误: %v", imageName, os, arch, err),
		Err: err,
	}
}

// NewImageNotFoundError 创建镜像不存在错误
func NewImageNotFoundError(imageName string, err error) *DiptError {
	return &DiptError{
		Type: ErrorImageNotFound,
		Message: fmt.Sprintf("镜像 %s 不存在\n建议：\n"+
			"1. 检查镜像名称和版本是否正确\n"+
			"2. 访问 https://hub.docker.com 验证镜像是否存在\n"+
			"3. 检查是否需要登录私有仓库\n"+
			"原始错误: %v", imageName, err),
		Err: err,
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(registry string, err error) *DiptError {
    return &DiptError{
        Type: ErrorUnauthorized,
        Message: fmt.Sprintf("访问镜像仓库 %s 未授权\n建议：\n"+
            "1. 检查 ~/.dipt_config 或 ./config.json 是否存在且格式正确\n"+
            "2. 验证用户名和密码是否正确\n"+
            "3. 确认是否有权限访问该镜像\n"+
            "原始错误: %v", registry, err),
        Err: err,
    }
}

// NewNetworkError 创建网络错误
func NewNetworkError(err error) *DiptError {
	return &DiptError{
		Type: ErrorNetwork,
		Message: fmt.Sprintf("网络连接错误\n建议：\n"+
			"1. 检查网络连接是否正常\n"+
			"2. 验证是否需要配置代理\n"+
			"3. 确认 DNS 解析是否正常\n"+
			"原始错误: %v", err),
		Err: err,
	}
}

// IsManifestUnknownError 检查是否是清单未找到错误
func IsManifestUnknownError(err error) bool {
    if err == nil {
        return false
    }
    var te *transport.Error
    if errors.As(err, &te) {
        for _, e := range te.Errors {
            if e.Code == transport.ManifestUnknownErrorCode {
                return true
            }
        }
    }
    return strings.Contains(strings.ToUpper(err.Error()), "MANIFEST_UNKNOWN")
}

// IsPlatformNotSupportedError 检查是否是平台不支持错误
func IsPlatformNotSupportedError(err error) bool {
    if err == nil {
        return false
    }
    // 常见于未匹配到指定平台的子清单
    if strings.Contains(strings.ToLower(err.Error()), "no child with platform") {
        return true
    }
    return IsManifestUnknownError(err)
}

// IsUnauthorizedError 检查是否是未授权错误
func IsUnauthorizedError(err error) bool {
    if err == nil {
        return false
    }
    var te *transport.Error
    if errors.As(err, &te) {
        for _, e := range te.Errors {
            if e.Code == transport.UnauthorizedErrorCode || e.Code == transport.DeniedErrorCode {
                return true
            }
        }
    }
    s := strings.ToUpper(err.Error())
    return strings.Contains(s, "UNAUTHORIZED") || strings.Contains(s, "DENIED")
}

// IsNetworkError 检查是否是网络错误
func IsNetworkError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "dial tcp") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "connection refused"))
}
