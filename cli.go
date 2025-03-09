package main

import (
	"fmt"
	"os"
	"runtime"
)

// parseArgs 解析命令行参数
func parseArgs() (imageName string, outputFile string, platform Platform, err error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return "", "", Platform{}, fmt.Errorf("用法: dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]")
	}

	// 设置默认值
	platform = Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
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

	// 如果没有指定输出文件，则根据镜像信息生成
	if outputFile == "" {
		outputFile = generateOutputFileName(imageName, platform)
	}

	return imageName, outputFile, platform, nil
}
