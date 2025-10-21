package color

import (
	"os"
	"testing"
)

func TestIsColorSupported(t *testing.T) {
	// 保存原始环境变量
	originalNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)
	
	// 测试 NO_COLOR 环境变量
	os.Setenv("NO_COLOR", "1")
	if isColorSupported() {
		t.Error("Expected color to be unsupported when NO_COLOR is set")
	}
	
	// 清除 NO_COLOR
	os.Unsetenv("NO_COLOR")
	if !isColorSupported() {
		t.Error("Expected color to be supported when NO_COLOR is not set")
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		text     string
		noColor  bool
		expected string
	}{
		{
			name:     "with color",
			color:    Red,
			text:     "error",
			noColor:  false,
			expected: Red + "error" + Reset,
		},
		{
			name:     "without color",
			color:    Red,
			text:     "error",
			noColor:  true,
			expected: "error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.noColor {
				os.Setenv("NO_COLOR", "1")
			} else {
				os.Unsetenv("NO_COLOR")
			}
			
			result := colorize(tt.color, tt.text)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorFunctions(t *testing.T) {
	// 测试颜色文本函数
	tests := []struct {
		name string
		fn   func(string) string
		text string
		color string
	}{
		{"RedText", RedText, "error", Red},
		{"GreenText", GreenText, "success", Green},
		{"YellowText", YellowText, "warning", Yellow},
		{"BlueText", BlueText, "info", Blue},
		{"CyanText", CyanText, "progress", Cyan},
	}
	
	// 确保颜色开启
	os.Unsetenv("NO_COLOR")
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			expected := tt.color + tt.text + Reset
			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}
