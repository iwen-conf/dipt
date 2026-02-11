package components

import (
	"fmt"
	"io"
	"strings"

	"dipt/internal/tui/theme"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuChoice 菜单选项
type MenuChoice int

const (
	MenuPull MenuChoice = iota
	MenuSettings
	MenuMirrors
	MenuQuit
)

// menuItem 菜单项
type menuItem struct {
	title string
	desc  string
	icon  string
}

func (i menuItem) Title() string       { return i.icon + "  " + i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// menuDelegate 自定义列表渲染
type menuDelegate struct{}

func (d menuDelegate) Height() int                             { return 2 }
func (d menuDelegate) Spacing() int                            { return 1 }
func (d menuDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(menuItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := "  " + i.Description()

	if index == m.Index() {
		title = theme.SelectedStyle.Render("▸ " + title)
		desc = lipgloss.NewStyle().Foreground(theme.ColorSecondary).Render(desc)
	} else {
		title = lipgloss.NewStyle().Foreground(theme.ColorText).Render("  " + title)
		desc = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// MenuModel 主菜单模型
type MenuModel struct {
	list   list.Model
	choice MenuChoice
	chosen bool
}

// MenuChosenMsg 菜单选择消息
type MenuChosenMsg struct {
	Choice MenuChoice
}

// NewMenuModel 创建主菜单
func NewMenuModel() MenuModel {
	items := []list.Item{
		menuItem{title: "拉取镜像", desc: "从 Docker Registry 拉取并保存镜像", icon: "📦"},
		menuItem{title: "设置", desc: "配置默认平台、保存目录等", icon: "⚙️"},
		menuItem{title: "镜像源管理", desc: "添加、删除、测试镜像加速器", icon: "🔗"},
		menuItem{title: "退出", desc: "退出 DIPT", icon: "👋"},
	}

	l := list.New(items, menuDelegate{}, 50, 14)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return MenuModel{list: l}
}

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.choice = MenuChoice(m.list.Index())
			m.chosen = true
			return m, func() tea.Msg { return MenuChosenMsg{Choice: m.choice} }
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width - 4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MenuModel) View() string {
	var b strings.Builder
	b.WriteString(RenderLogo())
	b.WriteString(theme.SubtitleStyle.Render("  Docker 镜像拉取与保存工具"))
	b.WriteString("\n\n")
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(theme.HelpStyle.Render("  ↑↓ 选择 · enter 确认 · q 退出"))
	return b.String()
}
