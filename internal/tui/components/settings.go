package components

import (
	"fmt"
	"strings"

	"dipt/internal/config"
	"dipt/internal/tui/theme"
	"dipt/internal/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// settingsField 设置字段
type settingsField int

const (
	settingsOS settingsField = iota
	settingsArch
	settingsSaveDir
	settingsUsername
	settingsPassword
	settingsSave
)

// SettingsModel 设置视图
type SettingsModel struct {
	focused       settingsField
	osIdx         int
	archIdx       int
	dirInput      textinput.Model
	usernameInput textinput.Model
	passwordInput textinput.Model
	userConfig    *types.UserConfig
	message       string
	isError       bool
}

// NewSettingsModel 创建设置视图
func NewSettingsModel(cfg *types.UserConfig) SettingsModel {
	ti := textinput.New()
	ti.SetValue(cfg.DefaultSaveDir)
	ti.CharLimit = 256
	ti.Width = 50

	userInput := textinput.New()
	userInput.SetValue(cfg.Registry.Username)
	userInput.CharLimit = 128
	userInput.Width = 50
	userInput.Placeholder = "留空表示匿名访问"

	passInput := textinput.New()
	passInput.SetValue(cfg.Registry.Password)
	passInput.CharLimit = 128
	passInput.Width = 50
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '*'
	passInput.Placeholder = "留空表示匿名访问"

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

	return SettingsModel{
		osIdx:         osIdx,
		archIdx:       archIdx,
		dirInput:      ti,
		usernameInput: userInput,
		passwordInput: passInput,
		userConfig:    cfg,
	}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case "tab":
			m = m.nextField()
			return m, nil
		case "shift+tab":
			m = m.prevField()
			return m, nil
		case "left":
			if m.focused == settingsOS && m.osIdx > 0 {
				m.osIdx--
			} else if m.focused == settingsArch && m.archIdx > 0 {
				m.archIdx--
			}
			return m, nil
		case "right":
			if m.focused == settingsOS && m.osIdx < len(osOptions)-1 {
				m.osIdx++
			} else if m.focused == settingsArch && m.archIdx < len(archOptions)-1 {
				m.archIdx++
			}
			return m, nil
		case "enter":
			if m.focused == settingsSave {
				return m.save()
			}
			m = m.nextField()
			return m, nil
		}
	}

	if m.focused == settingsSaveDir {
		var cmd tea.Cmd
		m.dirInput, cmd = m.dirInput.Update(msg)
		return m, cmd
	}
	if m.focused == settingsUsername {
		var cmd tea.Cmd
		m.usernameInput, cmd = m.usernameInput.Update(msg)
		return m, cmd
	}
	if m.focused == settingsPassword {
		var cmd tea.Cmd
		m.passwordInput, cmd = m.passwordInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SettingsModel) nextField() SettingsModel {
	m.dirInput.Blur()
	m.usernameInput.Blur()
	m.passwordInput.Blur()
	if m.focused < settingsSave {
		m.focused++
	}
	switch m.focused {
	case settingsSaveDir:
		m.dirInput.Focus()
	case settingsUsername:
		m.usernameInput.Focus()
	case settingsPassword:
		m.passwordInput.Focus()
	}
	return m
}

func (m SettingsModel) prevField() SettingsModel {
	m.dirInput.Blur()
	m.usernameInput.Blur()
	m.passwordInput.Blur()
	if m.focused > 0 {
		m.focused--
	}
	switch m.focused {
	case settingsSaveDir:
		m.dirInput.Focus()
	case settingsUsername:
		m.usernameInput.Focus()
	case settingsPassword:
		m.passwordInput.Focus()
	}
	return m
}

func (m SettingsModel) save() (SettingsModel, tea.Cmd) {
	m.userConfig.DefaultOS = osOptions[m.osIdx]
	m.userConfig.DefaultArch = archOptions[m.archIdx]
	dir := strings.TrimSpace(m.dirInput.Value())
	if dir != "" {
		m.userConfig.DefaultSaveDir = dir
	}
	m.userConfig.Registry.Username = strings.TrimSpace(m.usernameInput.Value())
	m.userConfig.Registry.Password = strings.TrimSpace(m.passwordInput.Value())

	if err := config.SaveUserConfig(m.userConfig); err != nil {
		m.message = "保存失败: " + err.Error()
		m.isError = true
	} else {
		m.message = "设置已保存"
		m.isError = false
	}
	return m, nil
}

func (m SettingsModel) View() string {
	var b strings.Builder
	b.WriteString(theme.TitleStyle.Render("  设置"))
	b.WriteString("\n\n")

	// OS
	b.WriteString("  默认操作系统: ")
	for i, opt := range osOptions {
		if i == m.osIdx {
			if m.focused == settingsOS {
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

	// Arch
	b.WriteString("  默认架构:     ")
	for i, opt := range archOptions {
		if i == m.archIdx {
			if m.focused == settingsArch {
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

	// Save dir
	label := "  保存目录:     "
	if m.focused == settingsSaveDir {
		label = theme.HighlightStyle.Render(label)
	}
	b.WriteString(label + m.dirInput.View() + "\n\n")

	// Username
	userLabel := "  Registry 用户: "
	if m.focused == settingsUsername {
		userLabel = theme.HighlightStyle.Render(userLabel)
	}
	b.WriteString(userLabel + m.usernameInput.View() + "\n\n")

	// Password
	passLabel := "  Registry 密码: "
	if m.focused == settingsPassword {
		passLabel = theme.HighlightStyle.Render(passLabel)
	}
	b.WriteString(passLabel + m.passwordInput.View() + "\n\n")

	// Save button
	if m.focused == settingsSave {
		b.WriteString("  " + theme.SelectedStyle.Render("[ 保存设置 ]"))
	} else {
		b.WriteString("  [ 保存设置 ]")
	}

	if m.message != "" {
		b.WriteString("\n\n")
		if m.isError {
			b.WriteString("  " + theme.ErrorStyle.Render(m.message))
		} else {
			b.WriteString("  " + theme.SuccessStyle.Render(m.message))
		}
	}

	b.WriteString("\n\n" + theme.HelpStyle.Render("  tab 切换 · ←→ 选择 · enter 确认 · esc 返回"))
	return b.String()
}
