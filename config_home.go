package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UserConfig 用户配置结构
type UserConfig struct {
	DefaultOS      string `json:"default_os"`       // 默认操作系统
	DefaultArch    string `json:"default_arch"`     // 默认架构
	DefaultSaveDir string `json:"default_save_dir"` // 默认保存目录
}

const configFileName = ".dipt_config"

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %v", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// loadUserConfig 加载用户配置
func loadUserConfig() (*UserConfig, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	// 如果配置文件不存在，启动交互式配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("未检测到配置文件 %s\n", configPath)
		return interactiveConfig()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// saveUserConfig 保存用户配置
func saveUserConfig(config *UserConfig) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// setConfigValue 设置配置值
func setConfigValue(key, value string) error {
	config, err := loadUserConfig()
	if err != nil {
		return err
	}

	switch key {
	case "os":
		config.DefaultOS = value
	case "arch":
		config.DefaultArch = value
	case "save_dir":
		// 验证目录是否存在
		if _, err := os.Stat(value); os.IsNotExist(err) {
			// 尝试创建目录
			err = os.MkdirAll(value, 0755)
			if err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		}
		// 转换为绝对路径
		absPath, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("转换路径失败: %v", err)
		}
		config.DefaultSaveDir = absPath
	default:
		return fmt.Errorf("未知的配置项: %s", key)
	}

	return saveUserConfig(config)
}
