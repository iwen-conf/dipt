package components

import (
	"fmt"
	"strings"

	"dipt/internal/tui/theme"
	"dipt/internal/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// StartPullMsg 开始拉取消息
type StartPullMsg struct {
	ImageName  string
	OutputFile string
	Platform   types.Platform
}

// BackToMenuMsg 返回菜单消息
type BackToMenuMsg struct{}

// pullFormField 表单字段索引
type pullFormField int

const (
	fieldImage pullFormField = iota
	fieldOutput
	fieldOS
	fieldArch
)

// PullFormModel 拉取表单
type PullFormModel struct {
	imageInput  textinput.Model
	outputInput textinput.Model
	osIdx       int
	archIdx     int
	focused     pullFormField
	userConfig  *types.UserConfig
	err         string
}

// NewPullFormModel 创建拉取表单
func NewPullFormModel(cfg *types.UserConfig) PullFormModel {
	imgInput := textinput.New()
	imgInput.Placeholder = "nginx:latest"
	imgInput.CharLimit = 256
	imgInput.Width = 50
	imgInput.Focus()

	outInput := textinput.New()
	outInput.Placeholder = "留空则自动生成"
	outInput.CharLimit = 256
	outInput.Width = 50

	// 根据用户配置设置默认 OS/Arch 索引
	osIdx := 0
	for i, o := range osOptions {
		if o == cfg.DefaultOS {
			osIdx = i
			break
		}
	}
	archIdx := 0
	for i, a := range archOptions {
		if a == cfg.DefaultArch {
			archIdx = i
			break
		}
	}

	return PullFormModel{
		imageInput:  imgInput,
		outputInput: outInput,
		osIdx:       osIdx,
		archIdx:     archIdx,
		userConfig:  cfg,
	}
}

func (m PullFormModel) Init() tea.Cmd { return textinput.Blink }

func (m PullFormModel) Update(msg tea.Msg) (PullFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case "tab", "shift+tab":
			return m.cycleFocus(msg.String() == "shift+tab"), nil
		case "enter":
			if m.focused == fieldArch {
				return m.submit()
			}
			return m.cycleFocus(false), nil
		case "left":
			if m.focused == fieldOS && m.osIdx > 0 {
				m.osIdx--
			} else if m.focused == fieldArch && m.archIdx > 0 {
				m.archIdx--
			}
			return m, nil
		case "right":
			if m.focused == fieldOS && m.osIdx < len(osOptions)-1 {
				m.osIdx++
			} else if m.focused == fieldArch && m.archIdx < len(archOptions)-1 {
				m.archIdx++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case fieldImage:
		m.imageInput, cmd = m.imageInput.Update(msg)
	case fieldOutput:
		m.outputInput, cmd = m.outputInput.Update(msg)
	}
	return m, cmd
}

func (m PullFormModel) cycleFocus(reverse bool) PullFormModel {
	if reverse {
		if m.focused > 0 {
			m.focused--
		}
	} else {
		if m.focused < fieldArch {
			m.focused++
		}
	}
	m.imageInput.Blur()
	m.outputInput.Blur()
	switch m.focused {
	case fieldImage:
		m.imageInput.Focus()
	case fieldOutput:
		m.outputInput.Focus()
	}
	return m
}

func (m PullFormModel) submit() (PullFormModel, tea.Cmd) {
	imageName := strings.TrimSpace(m.imageInput.Value())
	if imageName == "" {
		m.err = "请输入镜像名称"
		return m, nil
	}
	m.err = ""

	outputFile := strings.TrimSpace(m.outputInput.Value())

	return m, func() tea.Msg {
		return StartPullMsg{
			ImageName:  imageName,
			OutputFile: outputFile,
			Platform: types.Platform{
				OS:   osOptions[m.osIdx],
				Arch: archOptions[m.archIdx],
			},
		}
	}
}

func (m PullFormModel) View() string {
	var b strings.Builder
	b.WriteString(theme.TitleStyle.Render("  拉取镜像"))
	b.WriteString("\n\n")

	// 镜像名称
	label := "  镜像名称: "
	if m.focused == fieldImage {
		label = theme.HighlightStyle.Render(label)
	}
	b.WriteString(label + m.imageInput.View() + "\n\n")

	// 输出文件
	label = "  输出文件: "
	if m.focused == fieldOutput {
		label = theme.HighlightStyle.Render(label)
	}
	b.WriteString(label + m.outputInput.View() + "\n\n")

	// 操作系统选择
	b.WriteString("  操作系统: ")
	for i, opt := range osOptions {
		if i == m.osIdx {
			if m.focused == fieldOS {
				b.WriteString(theme.SelectedStyle.Render("[" + opt + "]"))
			} else {
				b.WriteString(theme.HighlightStyle.Render("[" + opt + "]"))
			}
		} else {
			b.WriteString(fmt.Sprintf(" %s ", opt))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// 架构选择
	b.WriteString("  架构:     ")
	for i, opt := range archOptions {
		if i == m.archIdx {
			if m.focused == fieldArch {
				b.WriteString(theme.SelectedStyle.Render("[" + opt + "]"))
			} else {
				b.WriteString(theme.HighlightStyle.Render("[" + opt + "]"))
			}
		} else {
			b.WriteString(fmt.Sprintf(" %s ", opt))
		}
		b.WriteString(" ")
	}

	if m.err != "" {
		b.WriteString("\n\n" + theme.ErrorStyle.Render("  "+m.err))
	}

	b.WriteString("\n\n" + theme.HelpStyle.Render("  tab 切换字段 · ←→ 选择平台 · enter 开始拉取 · esc 返回"))
	return b.String()
}
