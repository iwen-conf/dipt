package types

// Config 定义 JSON 配置文件的结构
type Config struct {
	Registry Registry `json:"registry"`
}

// Registry 镜像仓库配置
type Registry struct {
	Mirrors  []string `json:"mirrors,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}

// Platform 定义平台信息
type Platform struct {
	OS   string
	Arch string
}

// UserConfig 用户配置结构
type UserConfig struct {
	DefaultOS      string   `json:"default_os"`       // 默认操作系统
	DefaultArch    string   `json:"default_arch"`     // 默认架构
	DefaultSaveDir string   `json:"default_save_dir"` // 默认保存目录
	Registry       Registry `json:"registry"`         // 镜像仓库配置
}
