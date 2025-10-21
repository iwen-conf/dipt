package logger

import (
	"fmt"
	"sync"
	"time"

	"dipt/internal/color"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelSuccess
)

type Logger struct {
	mu       sync.Mutex
	verbose  bool
	colorOut bool
	stats    Stats
}

type Stats struct {
	Successes int
	Warnings  int
	Errors    int
	StartTime time.Time
	EndTime   time.Time
}

var instance *Logger
var once sync.Once

// GetLogger 获取单例 logger
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			verbose:  false,
			colorOut: true,
			stats: Stats{
				StartTime: time.Now(),
			},
		}
	})
	return instance
}

// Init 初始化日志系统
func (l *Logger) Init(verbose bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.verbose = verbose
	
	// 禁用日志文件写入功能
	// 仅使用终端输出
	
	return nil
}

// Close 关闭日志系统
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.stats.EndTime.IsZero() {
		l.stats.EndTime = time.Now()
	}
}

// SetVerbose 设置详细模式
func (l *Logger) SetVerbose(verbose bool) {
	l.verbose = verbose
}

// logToFile 写入日志到文件（已禁用）
func (l *Logger) logToFile(level LogLevel, message string) {
	// 日志文件写入功能已禁用
	// 仅使用终端彩色输出
	return
}

// Debug 输出调试信息
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelDebug, message)
	
	if l.verbose {
		color.PrintInfo("[DEBUG] " + message)
	}
}

// Info 输出信息
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelInfo, message)
	color.PrintInfo(message)
}

// Warning 输出警告
func (l *Logger) Warning(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Warnings++
	l.mu.Unlock()
	
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelWarning, message)
	color.PrintWarning(message)
}

// Error 输出错误
func (l *Logger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Errors++
	l.mu.Unlock()
	
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelError, message)
	color.PrintError(message)
}

// Success 输出成功信息
func (l *Logger) Success(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Successes++
	l.mu.Unlock()
	
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelSuccess, message)
	color.PrintSuccess(message)
}

// Progress 输出进度信息
func (l *Logger) Progress(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logToFile(LevelInfo, message)
	color.Progress(format, args...)
}

// Stage 输出阶段信息
func (l *Logger) Stage(stage string) {
	l.logToFile(LevelInfo, "=== " + stage + " ===")
	color.PrintStage(stage)
}

// GetStats 获取统计信息
func (l *Logger) GetStats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stats
}

// PrintSummary 打印总结报告
func (l *Logger) PrintSummary() {
	l.mu.Lock()
	if l.stats.EndTime.IsZero() {
		l.stats.EndTime = time.Now()
	}
	l.mu.Unlock()
	
	stats := l.GetStats()
	duration := stats.EndTime.Sub(stats.StartTime)
	
	color.PrintStage("执行总结")
	
	fmt.Println()
	fmt.Printf("✅ 修复成功项: %s\n", color.GreenText(fmt.Sprintf("%d", stats.Successes)))
	fmt.Printf("⚠️  检测警告项: %s\n", color.YellowText(fmt.Sprintf("%d", stats.Warnings)))
	fmt.Printf("❌ 未修复错误: %s\n", color.RedText(fmt.Sprintf("%d", stats.Errors)))
	fmt.Printf("⏱️  总耗时: %s\n", color.CyanText(duration.Round(time.Millisecond).String()))
	fmt.Println()
	
	if stats.Errors > 0 {
		color.PrintWarning("部分任务执行失败")
	} else if stats.Warnings > 0 {
		color.PrintInfo("任务完成，但存在一些警告")
	} else {
		color.PrintSuccess("所有任务执行成功！")
	}
}

// SaveReport 保存报告到文件（已禁用）
func (l *Logger) SaveReport(filename string) error {
	// 报告保存功能已禁用
	// 仅在终端显示总结信息
	return nil
}
