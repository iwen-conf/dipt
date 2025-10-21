package retry

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	
	"dipt/internal/logger"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数
	InitialBackoff time.Duration // 初始退避时间
	MaxBackoff     time.Duration // 最大退避时间
	BackoffFactor  float64       // 退避因子
	Jitter         bool          // 是否添加抖动
}

// DefaultConfig 默认重试配置
func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         true,
	}
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// WithRetry 执行带重试的函数
func WithRetry(fn RetryableFunc, config RetryConfig, operationName string) error {
	log := logger.GetLogger()
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := calculateBackoff(attempt, config)
			log.Info("第 %d 次重试 %s，等待 %v...", attempt, operationName, backoff)
			time.Sleep(backoff)
		}
		
		err := fn()
		if err == nil {
			if attempt > 0 {
				log.Success("%s 在第 %d 次重试后成功", operationName, attempt)
			}
			return nil
		}
		
		lastErr = err
		
		if attempt < config.MaxRetries {
			log.Warning("%s 失败: %v，准备重试...", operationName, err)
		} else {
			log.Error("%s 在 %d 次重试后仍然失败", operationName, config.MaxRetries)
		}
	}
	
	return fmt.Errorf("%s 失败，已重试 %d 次: %w", operationName, config.MaxRetries, lastErr)
}

// calculateBackoff 计算退避时间
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// 指数退避
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffFactor, float64(attempt-1))
	
	// 限制最大退避时间
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}
	
	// 添加抖动以避免雷鸣群问题
	if config.Jitter {
		jitter := rand.Float64() * 0.3 * backoff // 最多30%的抖动
		if rand.Intn(2) == 0 {
			backoff += jitter
		} else {
			backoff -= jitter
		}
	}
	
	return time.Duration(backoff)
}

// RetryableOperation 可重试的操作接口
type RetryableOperation interface {
	Execute() error
	GetName() string
	ShouldRetry(error) bool
}

// ExecuteWithRetry 执行可重试的操作
func ExecuteWithRetry(operation RetryableOperation, config RetryConfig) error {
	log := logger.GetLogger()
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := calculateBackoff(attempt, config)
			log.Info("重试 %s (第 %d/%d 次)，等待 %v...", 
				operation.GetName(), attempt, config.MaxRetries, backoff)
			time.Sleep(backoff)
		}
		
		err := operation.Execute()
		if err == nil {
			if attempt > 0 {
				log.Success("%s 在第 %d 次重试后成功", operation.GetName(), attempt)
			}
			return nil
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if !operation.ShouldRetry(err) {
			log.Error("%s 失败且不可重试: %v", operation.GetName(), err)
			return err
		}
		
		if attempt < config.MaxRetries {
			log.Warning("%s 失败 (第 %d 次): %v", operation.GetName(), attempt+1, err)
		}
	}
	
	return fmt.Errorf("%s 在 %d 次重试后失败: %w", 
		operation.GetName(), config.MaxRetries, lastErr)
}
