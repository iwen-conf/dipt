package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// parseArgs 解析命令行参数
func parseArgs() (imageName string, outputFile string, platform Platform, err error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return "", "", Platform{}, fmt.Errorf("用法:\n" +
			"拉取镜像: dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]\n" +
			"设置默认值: dipt set <os|arch|save_dir> <值>")
	}

	// 处理配置命令
	if args[0] == "set" {
		if len(args) != 3 {
			return "", "", Platform{}, fmt.Errorf("设置配置的用法: dipt set <os|arch|save_dir> <值>")
		}
		err := setConfigValue(args[1], args[2])
		if err != nil {
			return "", "", Platform{}, err
		}
		fmt.Printf("✅ 已设置 %s = %s\n", args[1], args[2])
		os.Exit(0)
	}

	// 加载用户配置
	userConfig, err := loadUserConfig()
	if err != nil {
		return "", "", Platform{}, fmt.Errorf("加载用户配置失败: %v", err)
	}

	// 设置默认值
	platform = Platform{
		OS:   userConfig.DefaultOS,
		Arch: userConfig.DefaultArch,
	}

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-os":
			if i+1 >= len(args) {
				return "", "", Platform{}, fmt.Errorf("-os 参数需要指定系统名称")
			}
			platform.OS = args[i+1]
			i++
		case "-arch":
			if i+1 >= len(args) {
				return "", "", Platform{}, fmt.Errorf("-arch 参数需要指定架构名称")
			}
			platform.Arch = args[i+1]
			i++
		default:
			// 如果不是选项参数，则认为是镜像名称或输出文件
			if imageName == "" {
				imageName = args[i]
			} else {
				outputFile = args[i]
			}
		}
	}

	if imageName == "" {
		return "", "", Platform{}, fmt.Errorf("必须指定镜像名称")
	}

	// 如果没有指定输出文件，则根据镜像信息生成并放在默认保存目录
	if outputFile == "" {
		outputFile = generateOutputFileName(imageName, platform)
		outputFile = filepath.Join(userConfig.DefaultSaveDir, outputFile)
	}

	return imageName, outputFile, platform, nil
}
