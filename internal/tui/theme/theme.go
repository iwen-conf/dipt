package theme

import "github.com/charmbracelet/lipgloss"

// 渐变色定义（紫->蓝->青）
var GradientColors = []string{
	"#9B59B6", "#8E44AD", "#6C5CE7", "#4A69BD",
	"#3498DB", "#2980B9", "#1ABC9C", "#00CEC9",
}

// 基础颜色
var (
	ColorPrimary   = lipgloss.Color("#6C5CE7")
	ColorSecondary = lipgloss.Color("#00CEC9")
	ColorSuccess   = lipgloss.Color("#00B894")
	ColorWarning   = lipgloss.Color("#FDCB6E")
	ColorError     = lipgloss.Color("#E17055")
	ColorMuted     = lipgloss.Color("#636E72")
	ColorText      = lipgloss.Color("#DFE6E9")
	ColorBg        = lipgloss.Color("#2D3436")
)

// 全局样式
var (
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF"))

	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	LogDebugStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	LogInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF"))
	LogWarningStyle = lipgloss.NewStyle().Foreground(ColorWarning)
	LogErrorStyle   = lipgloss.NewStyle().Foreground(ColorError)
	LogSuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
)

// GradientText 对文本应用渐变色
func GradientText(text string) string {
	if len(text) == 0 {
		return ""
	}
	runes := []rune(text)
	result := ""
	for i, r := range runes {
		colorIdx := i * len(GradientColors) / len(runes)
		if colorIdx >= len(GradientColors) {
			colorIdx = len(GradientColors) - 1
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(GradientColors[colorIdx]))
		result += style.Render(string(r))
	}
	return result
}
