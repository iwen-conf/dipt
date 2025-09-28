package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

// parseArgs 解析命令行参数（使用传入的默认配置用于缺省值）
func parseArgs(defaults *UserConfig) (imageName string, outputFile string, platform Platform, err error) {
    args := os.Args[1:]
    if len(args) == 0 {
        return "", "", Platform{}, fmt.Errorf("用法:\n" +
            "拉取镜像: dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]\n" +
            "设置默认值: dipt set <os|arch|save_dir> <值>\n" +
            "生成配置模板: dipt -conf new\n" +
            "镜像加速器管理:\n" +
            "  dipt mirror list          # 列出所有镜像加速器\n" +
            "  dipt mirror add <URL>     # 添加镜像加速器\n" +
            "  dipt mirror del <URL>     # 删除镜像加速器\n" +
            "  dipt mirror clear         # 清空所有镜像加速器")
    }

	// 处理生成配置模板命令
	if len(args) == 2 && args[0] == "-conf" && args[1] == "new" {
		err := generateConfigTemplate()
		if err != nil {
			return "", "", Platform{}, err
		}
		os.Exit(0)
	}

	// 处理镜像加速器命令
	if args[0] == "mirror" {
		err := handleMirrorCommand(args[1:])
		if err != nil {
			return "", "", Platform{}, err
		}
		os.Exit(0)
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

    // 设置默认值
    platform = Platform{
        OS:   defaults.DefaultOS,
        Arch: defaults.DefaultArch,
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
        outputFile = filepath.Join(defaults.DefaultSaveDir, outputFile)
    }

    return imageName, outputFile, platform, nil
}

// generateConfigTemplate 生成配置文件模板
func generateConfigTemplate() error {
	config := Config{}
	config.Registry.Mirrors = []string{
		"https://registry.docker-cn.com",
		"https://docker.mirrors.ustc.edu.cn",
		"http://hub-mirror.c.163.com",
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 检查当前目录下是否已存在配置文件
	if _, err := os.Stat("config.json"); err == nil {
		return fmt.Errorf("配置文件已存在，请先备份或删除现有的 config.json")
	}

	// 写入配置文件
	err = os.WriteFile("config.json", data, 0644)
	if err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	fmt.Println("✅ 配置模板已生成：config.json")
	fmt.Println("💡 提示：")
	fmt.Println("1. 您可以编辑配置文件添加认证信息")
	fmt.Println("2. 如果不需要认证，可以保持为空")
	fmt.Println("3. mirrors 字段用于配置镜像加速器")
	fmt.Println("4. 您也可以使用 'dipt mirror' 命令管理镜像加速器")
	return nil
}
