package main

// Config 定义 JSON 配置文件的结构
type Config struct {
	Registry struct {
		Mirrors  []string `json:"mirrors,omitempty"`
		Username string   `json:"username,omitempty"`
		Password string   `json:"password,omitempty"`
	} `json:"registry"`
}

// Platform 定义平台信息
type Platform struct {
	OS   string
	Arch string
}
