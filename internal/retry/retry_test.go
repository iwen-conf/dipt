package retry

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		attempts := 0
		err := WithRetry(func() error {
			attempts++
			return nil
		}, DefaultConfig(), "test operation")
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})
	
	t.Run("succeeds after retry", func(t *testing.T) {
		attempts := 0
		err := WithRetry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		}, DefaultConfig(), "test operation")
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})
	
	t.Run("fails after max retries", func(t *testing.T) {
		attempts := 0
		config := RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			BackoffFactor:  2.0,
			Jitter:         false,
		}
		
		err := WithRetry(func() error {
			attempts++
			return errors.New("persistent error")
		}, config, "test operation")
		
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 3 { // 1 initial + 2 retries
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
		BackoffFactor:  2.0,
		Jitter:         false,
	}
	
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 10 * time.Second}, // Capped at max
		{6, 10 * time.Second}, // Still capped
	}
	
	for _, tt := range tests {
		t.Run(string(rune(tt.attempt)), func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, config)
			if backoff != tt.expected {
				t.Errorf("Expected backoff %v for attempt %d, got %v",
					tt.expected, tt.attempt, backoff)
			}
		})
	}
}

func TestRetryableOperation(t *testing.T) {
	t.Run("operation with custom retry logic", func(t *testing.T) {
		op := &testOperation{
			name:      "test op",
			maxFails:  2,
			failCount: 0,
		}
		
		config := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			BackoffFactor:  2.0,
			Jitter:         false,
		}
		
		err := ExecuteWithRetry(op, config)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
		if op.failCount != 2 {
			t.Errorf("Expected 2 failures before success, got %d", op.failCount)
		}
	})
}

// testOperation implements RetryableOperation for testing
type testOperation struct {
	name      string
	maxFails  int
	failCount int
}

func (o *testOperation) Execute() error {
	if o.failCount < o.maxFails {
		o.failCount++
		return errors.New("temporary failure")
	}
	return nil
}

func (o *testOperation) GetName() string {
	return o.name
}

func (o *testOperation) ShouldRetry(err error) bool {
	// Always retry in this test
	return true
}
