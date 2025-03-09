package main

import (
	"fmt"
	"os"
)

func main() {
	imageName, outputFile, platform, err := parseArgs()
	if err != nil {
		fmt.Println("错误:", err)
		fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
		os.Exit(1)
	}

	// 加载配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	// 拉取镜像并保存
	fmt.Printf("正在拉取镜像 %s (系统: %s, 架构: %s)...\n", imageName, platform.OS, platform.Arch)
	err = pullAndSaveImage(imageName, outputFile, platform, config)
	if err != nil {
		if diptErr, ok := err.(*DiptError); ok {
			// 使用自定义错误的格式化信息
			fmt.Println("\n❌ 错误:", diptErr.Message)
		} else {
			fmt.Println("\n❌ 错误:", err)
		}
		os.Exit(1)
	}

	fmt.Printf("\n✅ 镜像已保存到 %s\n", outputFile)
}
