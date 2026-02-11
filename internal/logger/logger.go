package logger

import (
	"fmt"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelSuccess
)

// Logger 日志器（TUI 兼容）
type Logger struct {
	mu      sync.Mutex
	verbose bool
	stats   Stats
}

// Stats 统计信息
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

// Debug 输出调试信息（TUI 模式下静默）
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Info 输出信息（TUI 模式下静默）
func (l *Logger) Info(format string, args ...interface{}) {
	// TUI 模式下不直接输出到终端
}

// Warning 输出警告
func (l *Logger) Warning(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Warnings++
	l.mu.Unlock()
}

// Error 输出错误
func (l *Logger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Errors++
	l.mu.Unlock()
}

// Success 输出成功信息
func (l *Logger) Success(format string, args ...interface{}) {
	l.mu.Lock()
	l.stats.Successes++
	l.mu.Unlock()
}

// Progress 输出进度信息（TUI 模式下静默）
func (l *Logger) Progress(format string, args ...interface{}) {}

// Stage 输出阶段信息（TUI 模式下静默）
func (l *Logger) Stage(stage string) {}

// GetStats 获取统计信息
func (l *Logger) GetStats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stats
}

// PrintSummary TUI 模式下不打印总结
func (l *Logger) PrintSummary() {}

// SaveReport 保存报告（已禁用）
func (l *Logger) SaveReport(filename string) error {
	return nil
}
