package main

// Config 定义 JSON 配置文件的结构
type Config struct {
	Registry struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"registry"`
}

// Platform 定义平台信息
type Platform struct {
	OS   string
	Arch string
}
