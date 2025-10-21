package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dipt/internal/types"
)

// InteractiveConfig 交互式配置
func InteractiveConfig() (*types.UserConfig, error) {
	fmt.Println("👋 欢迎使用 DIPT！")
	fmt.Println("📝 首次运行需要进行一些基本设置...")

	reader := bufio.NewReader(os.Stdin)
	config := &types.UserConfig{}

	// 设置默认操作系统
	fmt.Printf("\n💻 请选择默认的操作系统 [linux/windows/darwin] (默认: linux): ")
	osName, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取输入失败: %v", err)
	}
	osName = strings.TrimSpace(osName)
	if osName == "" {
		osName = "linux"
	}
	if !isValidOS(osName) {
		return nil, fmt.Errorf("不支持的操作系统: %s", osName)
	}
	config.DefaultOS = osName

	// 设置默认架构
	fmt.Printf("🔧 请选择默认的架构 [amd64/arm64/arm/386] (默认: amd64): ")
	arch, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取输入失败: %v", err)
	}
	arch = strings.TrimSpace(arch)
	if arch == "" {
		arch = "amd64"
	}
	if !isValidArch(arch) {
		return nil, fmt.Errorf("不支持的架构: %s", arch)
	}
	config.DefaultArch = arch

	// 设置默认保存目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %v", err)
	}
	defaultSaveDir := filepath.Join(homeDir, "DockerImages")

	fmt.Printf("📂 请输入默认的镜像保存目录 (默认: %s): ", defaultSaveDir)
	saveDir, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取输入失败: %v", err)
	}
	saveDir = strings.TrimSpace(saveDir)
	if saveDir == "" {
		saveDir = defaultSaveDir
	}

	// 创建保存目录
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			return nil, fmt.Errorf("创建保存目录失败: %v", err)
		}
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(saveDir)
	if err != nil {
		return nil, fmt.Errorf("转换路径失败: %v", err)
	}
	config.DefaultSaveDir = absPath

	// 保存配置
	if err := SaveUserConfig(config); err != nil {
		return nil, fmt.Errorf("保存配置失败: %v", err)
	}

	fmt.Printf("\n✅ 配置完成！配置文件已保存到: %s\n", configFileName)
	fmt.Println("💡 您可以随时使用 'dipt set' 命令修改这些设置")
	fmt.Println("   例如: dipt set os linux")
	fmt.Println("        dipt set arch arm64")
	fmt.Println("        dipt set save_dir ~/docker-images")
	fmt.Println()

	return config, nil
}

// isValidOS 检查操作系统是否有效
func isValidOS(os string) bool {
	validOS := []string{"linux", "windows", "darwin"}
	for _, v := range validOS {
		if v == os {
			return true
		}
	}
	return false
}

// isValidArch 检查架构是否有效
func isValidArch(arch string) bool {
	validArch := []string{"amd64", "arm64", "arm", "386"}
	for _, v := range validArch {
		if v == arch {
			return true
		}
	}
	return false
}
