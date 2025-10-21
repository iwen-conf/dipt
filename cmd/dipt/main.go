package main

import (
    "fmt"
    "os"
    "path/filepath"
    
    "dipt/internal/cli"
    "dipt/internal/color"
    "dipt/internal/config"
    "dipt/internal/errors"
    "dipt/internal/logger"
    "dipt/pkg/docker"
)

func main() {
    // 初始化日志系统
    log := logger.GetLogger()
    defer log.Close()
    
    // 先加载用户与项目配置（支持非交互环境与 env 覆盖）
    userConfig, effectiveRegistry, err := config.LoadEffectiveConfigs()
    if err != nil {
        color.Error("配置加载失败: %v", err)
        fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
        os.Exit(1)
    }
    
    // 解析命令行参数
    imageName, outputFile, platform, verbose, err := cli.ParseArgs(userConfig)
    if err != nil {
        color.Error("%v", err)
        fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
        os.Exit(1)
    }
    
    // 初始化日志详细模式
    if err := log.Init(verbose); err != nil {
        color.Error("日志系统初始化失败: %v", err)
    }
    log.SetVerbose(verbose)

    // 拉取镜像并保存
    log.Stage(fmt.Sprintf("拉取镜像 %s", imageName))
    log.Info("平台: %s/%s", platform.OS, platform.Arch)
    log.Info("输出文件: %s", outputFile)
    
    // 确保输出目录存在
    _ = os.MkdirAll(filepath.Dir(outputFile), 0755)
    
    err = docker.PullAndSaveImage(imageName, outputFile, platform, effectiveRegistry)
    if err != nil {
        if diptErr, ok := err.(*errors.DiptError); ok {
            // 使用自定义错误的格式化信息
            log.Error("%s", diptErr.Message)
        } else {
            log.Error("拉取镜像失败: %v", err)
        }
        
        // 打印总结
        log.PrintSummary()
        
        os.Exit(1)
    }
    
    log.Success("镜像已成功保存到 %s", outputFile)
    
    // 打印总结
    log.PrintSummary()
}
