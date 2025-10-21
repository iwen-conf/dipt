package color

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

// 检查是否支持颜色输出
func isColorSupported() bool {
	// 检查环境变量
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	// Windows 平台检查
	if runtime.GOOS == "windows" {
		// Windows 10 及以上支持 ANSI 颜色
		// 简化检查，假设现代 Windows 都支持
		return true
	}
	
	// Unix-like 系统默认支持
	return true
}

// 包装颜色输出（仅在支持颜色时添加颜色代码）
func colorize(color, text string) string {
	if !isColorSupported() {
		return text
	}
	return color + text + Reset
}

// Success 输出成功信息（绿色）
func Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Green, "✅ " + message))
}

// Error 输出错误信息（红色）
func Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Red, "❌ " + message))
}

// Warning 输出警告信息（黄色）
func Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Yellow, "⚠️ " + message))
}

// Info 输出信息（蓝色）
func Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Blue, "ℹ️ " + message))
}

// Progress 输出进度信息（青色）
func Progress(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Cyan, "⏳ " + message))
}

// Bold 加粗文本
func BoldText(text string) string {
	if !isColorSupported() {
		return text
	}
	return Bold + text + Reset
}

// 纯颜色文本输出（不带图标）
func RedText(text string) string {
	return colorize(Red, text)
}

func GreenText(text string) string {
	return colorize(Green, text)
}

func YellowText(text string) string {
	return colorize(Yellow, text)
}

func BlueText(text string) string {
	return colorize(Blue, text)
}

func CyanText(text string) string {
	return colorize(Cyan, text)
}

// PrintStage 打印阶段信息
func PrintStage(stage string) {
	fmt.Println(colorize(Purple+Bold, "\n━━━ " + stage + " ━━━"))
}

// PrintSuccess 打印成功结果（无格式化）
func PrintSuccess(message string) {
	fmt.Println(colorize(Green, "✅ " + message))
}

// PrintError 打印错误结果（无格式化）
func PrintError(message string) {
	fmt.Fprintln(os.Stderr, colorize(Red, "❌ " + message))
}

// PrintWarning 打印警告（无格式化）
func PrintWarning(message string) {
	fmt.Println(colorize(Yellow, "⚠️ " + message))
}

// PrintInfo 打印信息（无格式化）
func PrintInfo(message string) {
	fmt.Println(colorize(Blue, "ℹ️ " + message))
}
