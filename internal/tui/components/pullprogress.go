package components

import (
	"fmt"
	"strings"

	"dipt/internal/tui/theme"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressMsg 进度更新消息
type ProgressMsg struct {
	Downloaded int64
	Total      int64
}

// LogMsg 日志消息
type LogMsg struct {
	Level   string
	Message string
}

// PullDoneMsg 拉取完成消息
type PullDoneMsg struct {
	Err error
}

// PullProgressModel 拉取进度视图
type PullProgressModel struct {
	spinner    spinner.Model
	progress   progress.Model
	viewport   viewport.Model
	logs       []string
	downloaded int64
	total      int64
	done       bool
	err        error
	imageName  string
	width      int
}

// NewPullProgressModel 创建进度视图
func NewPullProgressModel(imageName string) PullProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorPrimary)

	p := progress.New(
		progress.WithGradient(string(theme.ColorPrimary), string(theme.ColorSecondary)),
		progress.WithWidth(50),
	)

	vp := viewport.New(60, 10)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorMuted).
		Padding(0, 1)

	return PullProgressModel{
		spinner:   s,
		progress:  p,
		viewport:  vp,
		imageName: imageName,
		width:     60,
	}
}

func (m PullProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m PullProgressModel) Update(msg tea.Msg) (PullProgressModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - 4
		m.viewport.Width = m.width
		m.progress = progress.New(
			progress.WithGradient(string(theme.ColorPrimary), string(theme.ColorSecondary)),
			progress.WithWidth(m.width-10),
		)
	case ProgressMsg:
		m.downloaded = msg.Downloaded
		m.total = msg.Total
		if m.total > 0 {
			pct := float64(m.downloaded) / float64(m.total)
			cmds = append(cmds, m.progress.SetPercent(pct))
		}
	case LogMsg:
		styled := m.styleLog(msg.Level, msg.Message)
		m.logs = append(m.logs, styled)
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
	case PullDoneMsg:
		m.done = true
		m.err = msg.Err
		if msg.Err != nil {
			m.logs = append(m.logs, theme.ErrorStyle.Render("拉取失败: "+msg.Err.Error()))
		} else {
			m.logs = append(m.logs, theme.SuccessStyle.Render("拉取完成!"))
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "enter", "esc":
				return m, func() tea.Msg { return BackToMenuMsg{} }
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m PullProgressModel) styleLog(level, message string) string {
	switch level {
	case "debug":
		return theme.LogDebugStyle.Render("[DEBUG] " + message)
	case "info":
		return theme.LogInfoStyle.Render("[INFO]  " + message)
	case "warning":
		return theme.LogWarningStyle.Render("[WARN]  " + message)
	case "error":
		return theme.LogErrorStyle.Render("[ERROR] " + message)
	case "success":
		return theme.LogSuccessStyle.Render("[OK]    " + message)
	default:
		return message
	}
}

func (m PullProgressModel) View() string {
	var b strings.Builder
	b.WriteString(theme.TitleStyle.Render("  拉取镜像"))
	b.WriteString("\n\n")

	if !m.done {
		b.WriteString(fmt.Sprintf("  %s 正在拉取 %s\n\n",
			m.spinner.View(),
			theme.HighlightStyle.Render(m.imageName)))
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("  %s 拉取失败\n\n",
			theme.ErrorStyle.Render("✗")))
	} else {
		b.WriteString(fmt.Sprintf("  %s 拉取完成\n\n",
			theme.SuccessStyle.Render("✓")))
	}

	// 进度条
	if m.total > 0 {
		b.WriteString("  " + m.progress.View() + "\n")
		b.WriteString(fmt.Sprintf("  %s / %s\n\n",
			formatBytes(m.downloaded), formatBytes(m.total)))
	}

	// 日志视图 — 需要对每行缩进，否则边框只有首行偏移
	b.WriteString("  " + strings.ReplaceAll(m.viewport.View(), "\n", "\n  ") + "\n")

	if m.done {
		b.WriteString("\n" + theme.HelpStyle.Render("  enter/esc 返回菜单"))
	}
	return b.String()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
