package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    // 先加载用户与项目配置（支持非交互环境与 env 覆盖）
    userConfig, effectiveRegistry, err := loadEffectiveConfigs()
    if err != nil {
        fmt.Println("错误:", err)
        fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
        os.Exit(1)
    }
    imageName, outputFile, platform, err := parseArgs(userConfig)
    if err != nil {
        fmt.Println("错误:", err)
        fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
        os.Exit(1)
    }

    // 拉取镜像并保存
    fmt.Printf("正在拉取镜像 %s (系统: %s, 架构: %s)...\n", imageName, platform.OS, platform.Arch)
    // 确保输出目录存在
    _ = os.MkdirAll(filepath.Dir(outputFile), 0755)
    err = pullAndSaveImage(imageName, outputFile, platform, effectiveRegistry)
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
