package cli

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "dipt/internal/config"
    "dipt/internal/types"
    "dipt/pkg/docker"
)

// CLIOptions 命令行选项
type CLIOptions struct {
    Verbose  bool   // 详细输出
    DryRun   bool   // 仅检测不执行
    Fix      bool   // 自动修复
    Mirror   string // 指定镜像源
}

// ParseArgs 解析命令行参数（使用传入的默认配置用于缺省值）
func ParseArgs(defaults *types.UserConfig) (imageName string, outputFile string, platform types.Platform, verbose bool, err error) {
    args := os.Args[1:]
    if len(args) == 0 {
        return "", "", types.Platform{}, false, fmt.Errorf("用法:\n" +
            "拉取镜像: dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]\n" +
            "设置默认值: dipt set <os|arch|save_dir> <值>\n" +
            "生成配置模板: dipt -conf new\n" +
            "镜像加速器管理:\n" +
            "  dipt mirror list          # 列出所有镜像加速器\n" +
            "  dipt mirror add <URL>     # 添加镜像加速器\n" +
            "  dipt mirror del <URL>     # 删除镜像加速器\n" +
            "  dipt mirror clear         # 清空所有镜像加速器\n" +
            "  dipt mirror test <URL>    # 测试镜像加速器\n" +
            "\n选项:\n" +
            "  --verbose                 # 显示详细日志\n" +
            "  --mirror=<URL>            # 指定镜像源\n" +
            "  --fix                     # 自动修复问题\n" +
            "  --dry-run                 # 仅检测不修改")
    }

	// 处理生成配置模板命令
	if len(args) == 2 && args[0] == "-conf" && args[1] == "new" {
		err := GenerateConfigTemplate()
		if err != nil {
			return "", "", types.Platform{}, false, err
		}
		os.Exit(0)
	}

	// 处理镜像加速器命令
	if args[0] == "mirror" {
		err := config.HandleMirrorCommand(args[1:])
		if err != nil {
			return "", "", types.Platform{}, false, err
		}
		os.Exit(0)
	}

	// 处理配置命令
	if args[0] == "set" {
		if len(args) != 3 {
			return "", "", types.Platform{}, false, fmt.Errorf("设置配置的用法: dipt set <os|arch|save_dir> <值>")
		}
		err := config.SetConfigValue(args[1], args[2])
		if err != nil {
			return "", "", types.Platform{}, false, err
		}
		fmt.Printf("✅ 已设置 %s = %s\n", args[1], args[2])
		os.Exit(0)
	}

    // 设置默认值
    platform = types.Platform{
        OS:   defaults.DefaultOS,
        Arch: defaults.DefaultArch,
    }

	// 解析参数
	var customMirror string
	var dryRun, fix bool
	
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-os":
			if i+1 >= len(args) {
				return "", "", types.Platform{}, false, fmt.Errorf("-os 参数需要指定系统名称")
			}
			platform.OS = args[i+1]
			i++
		case args[i] == "-arch":
			if i+1 >= len(args) {
				return "", "", types.Platform{}, false, fmt.Errorf("-arch 参数需要指定架构名称")
			}
			platform.Arch = args[i+1]
			i++
		case args[i] == "--verbose" || args[i] == "-v":
			verbose = true
		case args[i] == "--dry-run":
			dryRun = true
		case args[i] == "--fix":
			fix = true
		case strings.HasPrefix(args[i], "--mirror="):
			customMirror = strings.TrimPrefix(args[i], "--mirror=")
		case !strings.HasPrefix(args[i], "-"):
			// 如果不是选项参数，则认为是镜像名称或输出文件
			if imageName == "" {
				imageName = args[i]
			} else {
				outputFile = args[i]
			}
		}
	}
	
	// 处理特殊选项
	if customMirror != "" {
		// TODO: 在后续实现中使用自定义镜像源
		os.Setenv("DIPT_CUSTOM_MIRROR", customMirror)
	}
	if dryRun {
		os.Setenv("DIPT_DRY_RUN", "1")
	}
	if fix {
		os.Setenv("DIPT_AUTO_FIX", "1")
	}

	if imageName == "" {
		return "", "", types.Platform{}, false, fmt.Errorf("必须指定镜像名称")
	}

    // 如果没有指定输出文件，则根据镜像信息生成并放在默认保存目录
    if outputFile == "" {
        outputFile = docker.GenerateOutputFileName(imageName, platform)
        outputFile = filepath.Join(defaults.DefaultSaveDir, outputFile)
    }

    return imageName, outputFile, platform, verbose, nil
}

// GenerateConfigTemplate 生成配置文件模板
func GenerateConfigTemplate() error {
	config := types.Config{}
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
