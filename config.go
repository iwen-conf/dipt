package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// loadConfig 读取和解析 config.json 文件
func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return config, nil
}
