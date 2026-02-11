package components

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"dipt/internal/config"
	"dipt/internal/tui/theme"
	"dipt/internal/types"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MirrorTestResultMsg 镜像源测试结果
type MirrorTestResultMsg struct {
	URL       string
	Available bool
	Latency   time.Duration
	Err       error
}

// mirrorsMode 镜像源视图模式
type mirrorsMode int

const (
	mirrorsList mirrorsMode = iota
	mirrorsAdd
)

// MirrorsModel 镜像源管理视图
type MirrorsModel struct {
	table      table.Model
	addInput   textinput.Model
	mode       mirrorsMode
	userConfig *types.UserConfig
	message    string
	isError    bool
	testing    map[string]bool
}

// NewMirrorsModel 创建镜像源管理视图
func NewMirrorsModel(cfg *types.UserConfig) MirrorsModel {
	ti := textinput.New()
	ti.Placeholder = "https://mirror.example.com"
	ti.CharLimit = 256
	ti.Width = 50

	m := MirrorsModel{
		addInput:   ti,
		userConfig: cfg,
		testing:    make(map[string]bool),
	}
	m.table = m.buildTable()
	return m
}

func (m MirrorsModel) buildTable() table.Model {
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "镜像源 URL", Width: 45},
		{Title: "状态", Width: 12},
	}

	rows := make([]table.Row, len(m.userConfig.Registry.Mirrors))
	for i, mirror := range m.userConfig.Registry.Mirrors {
		status := "—"
		if m.testing[mirror] {
			status = "测试中..."
		}
		rows[i] = table.Row{fmt.Sprintf("%d", i+1), mirror, status}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.ColorMuted).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(theme.ColorPrimary).
		Bold(false)
	t.SetStyles(s)

	return t
}

func (m MirrorsModel) Init() tea.Cmd { return nil }

func (m MirrorsModel) Update(msg tea.Msg) (MirrorsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MirrorTestResultMsg:
		delete(m.testing, msg.URL)
		if msg.Available {
			m.message = fmt.Sprintf("%s 可用 (延迟: %v)", msg.URL, msg.Latency.Round(time.Millisecond))
			m.isError = false
		} else {
			errMsg := "未知错误"
			if msg.Err != nil {
				errMsg = msg.Err.Error()
			}
			m.message = fmt.Sprintf("%s 不可用: %s", msg.URL, errMsg)
			m.isError = true
		}
		m.table = m.buildTable()
		return m, nil

	case tea.KeyMsg:
		if m.mode == mirrorsAdd {
			return m.updateAddMode(msg)
		}
		return m.updateListMode(msg)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m MirrorsModel) updateListMode(msg tea.KeyMsg) (MirrorsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return BackToMenuMsg{} }
	case "a":
		m.mode = mirrorsAdd
		m.addInput.Focus()
		m.message = ""
		return m, nil
	case "d", "delete":
		return m.deleteCurrent()
	case "t":
		return m.testCurrent()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m MirrorsModel) updateAddMode(msg tea.KeyMsg) (MirrorsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = mirrorsList
		m.addInput.Blur()
		m.addInput.SetValue("")
		return m, nil
	case "enter":
		url := strings.TrimSpace(m.addInput.Value())
		if url == "" {
			return m, nil
		}
		// 检查重复
		for _, existing := range m.userConfig.Registry.Mirrors {
			if existing == url {
				m.message = "镜像源已存在"
				m.isError = true
				return m, nil
			}
		}
		m.userConfig.Registry.Mirrors = append(m.userConfig.Registry.Mirrors, url)
		if err := config.SaveUserConfig(m.userConfig); err != nil {
			m.message = "保存失败: " + err.Error()
			m.isError = true
		} else {
			m.message = "已添加: " + url
			m.isError = false
		}
		m.addInput.SetValue("")
		m.addInput.Blur()
		m.mode = mirrorsList
		m.table = m.buildTable()
		return m, nil
	}

	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m MirrorsModel) deleteCurrent() (MirrorsModel, tea.Cmd) {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.userConfig.Registry.Mirrors) {
		return m, nil
	}
	deleted := m.userConfig.Registry.Mirrors[idx]
	m.userConfig.Registry.Mirrors = append(
		m.userConfig.Registry.Mirrors[:idx],
		m.userConfig.Registry.Mirrors[idx+1:]...,
	)
	if err := config.SaveUserConfig(m.userConfig); err != nil {
		m.message = "删除失败: " + err.Error()
		m.isError = true
	} else {
		m.message = "已删除: " + deleted
		m.isError = false
	}
	m.table = m.buildTable()
	return m, nil
}

func (m MirrorsModel) testCurrent() (MirrorsModel, tea.Cmd) {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.userConfig.Registry.Mirrors) {
		return m, nil
	}
	url := m.userConfig.Registry.Mirrors[idx]
	m.testing[url] = true
	m.table = m.buildTable()
	m.message = "正在测试 " + url + " ..."
	m.isError = false

	return m, func() tea.Msg {
		return testMirrorCmd(url)
	}
}

func testMirrorCmd(url string) tea.Msg {
	mgr := newSimpleMirrorTester()
	available, latency, err := mgr.test(url)
	return MirrorTestResultMsg{
		URL:       url,
		Available: available,
		Latency:   latency,
		Err:       err,
	}
}

// simpleMirrorTester 简单镜像源测试器
type simpleMirrorTester struct{}

func newSimpleMirrorTester() simpleMirrorTester { return simpleMirrorTester{} }

func (t simpleMirrorTester) test(mirrorURL string) (bool, time.Duration, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	testURL := mirrorURL
	if !strings.HasPrefix(testURL, "http://") && !strings.HasPrefix(testURL, "https://") {
		testURL = "https://" + testURL
	}
	testURL = strings.TrimSuffix(testURL, "/") + "/v2/"

	start := time.Now()
	resp, err := client.Get(testURL)
	latency := time.Since(start)
	if err != nil {
		return false, latency, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 401 {
		return true, latency, nil
	}
	return false, latency, fmt.Errorf("状态码: %d", resp.StatusCode)
}

func (m MirrorsModel) View() string {
	var b strings.Builder
	b.WriteString(theme.TitleStyle.Render("  镜像源管理"))
	b.WriteString("\n\n")

	if len(m.userConfig.Registry.Mirrors) == 0 && m.mode != mirrorsAdd {
		b.WriteString("  暂无镜像源配置\n")
	} else if m.mode != mirrorsAdd {
		b.WriteString("  " + m.table.View() + "\n")
	}

	if m.mode == mirrorsAdd {
		b.WriteString("  添加镜像源:\n\n")
		b.WriteString("  " + m.addInput.View() + "\n")
		b.WriteString("\n" + theme.HelpStyle.Render("  enter 添加 · esc 取消"))
	} else {
		if m.message != "" {
			b.WriteString("\n")
			if m.isError {
				b.WriteString("  " + theme.ErrorStyle.Render(m.message))
			} else {
				b.WriteString("  " + theme.SuccessStyle.Render(m.message))
			}
		}
		b.WriteString("\n\n" + theme.HelpStyle.Render("  a 添加 · d 删除 · t 测试 · esc 返回"))
	}

	return b.String()
}

