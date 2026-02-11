package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dipt/internal/config"
	"dipt/internal/tui/theme"
	"dipt/internal/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SetupDoneMsg 配置完成消息
type SetupDoneMsg struct {
	Config *types.UserConfig
}

// setupStep 配置步骤
type setupStep int

const (
	stepOS setupStep = iota
	stepArch
	stepSaveDir
	stepConfirm
)

// SetupModel 首次运行配置向导
type SetupModel struct {
	step     setupStep
	osIdx    int
	archIdx  int
	dirInput textinput.Model
	err      string
}

var osOptions = []string{"linux", "windows", "darwin"}
var archOptions = []string{"amd64", "arm64", "arm", "386"}

// NewSetupModel 创建配置向导
func NewSetupModel() SetupModel {
	ti := textinput.New()
	homeDir, _ := os.UserHomeDir()
	ti.Placeholder = filepath.Join(homeDir, "DockerImages")
	ti.CharLimit = 256
	ti.Width = 50

	return SetupModel{
		step:     stepOS,
		dirInput: ti,
	}
}

func (m SetupModel) Init() tea.Cmd { return nil }

func (m SetupModel) Update(msg tea.Msg) (SetupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.step == stepOS && m.osIdx > 0 {
				m.osIdx--
			} else if m.step == stepArch && m.archIdx > 0 {
				m.archIdx--
			}
		case "down", "j":
			if m.step == stepOS && m.osIdx < len(osOptions)-1 {
				m.osIdx++
			} else if m.step == stepArch && m.archIdx < len(archOptions)-1 {
				m.archIdx++
			}
		case "enter":
			return m.handleEnter()
		}
	}

	if m.step == stepSaveDir {
		var cmd tea.Cmd
		m.dirInput, cmd = m.dirInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SetupModel) handleEnter() (SetupModel, tea.Cmd) {
	switch m.step {
	case stepOS:
		m.step = stepArch
	case stepArch:
		m.step = stepSaveDir
		m.dirInput.Focus()
	case stepSaveDir:
		m.step = stepConfirm
	case stepConfirm:
		return m, m.saveConfig
	}
	return m, nil
}

func (m SetupModel) saveConfig() tea.Msg {
	saveDir := m.dirInput.Value()
	if saveDir == "" {
		homeDir, _ := os.UserHomeDir()
		saveDir = filepath.Join(homeDir, "DockerImages")
	}
	absPath, err := filepath.Abs(saveDir)
	if err != nil {
		absPath = saveDir
	}
	_ = os.MkdirAll(absPath, 0755)

	cfg := &types.UserConfig{
		DefaultOS:      osOptions[m.osIdx],
		DefaultArch:    archOptions[m.archIdx],
		DefaultSaveDir: absPath,
	}
	_ = config.SaveUserConfig(cfg)
	return SetupDoneMsg{Config: cfg}
}

func (m SetupModel) View() string {
	var b strings.Builder
	b.WriteString(RenderLogo())
	b.WriteString(theme.TitleStyle.Render("  首次运行配置向导"))
	b.WriteString("\n\n")

	switch m.step {
	case stepOS:
		b.WriteString("  选择默认操作系统:\n\n")
		for i, opt := range osOptions {
			if i == m.osIdx {
				b.WriteString(fmt.Sprintf("  %s %s\n", theme.SelectedStyle.Render("▸"), theme.SelectedStyle.Render(opt)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", opt))
			}
		}
	case stepArch:
		b.WriteString(fmt.Sprintf("  操作系统: %s\n\n", theme.HighlightStyle.Render(osOptions[m.osIdx])))
		b.WriteString("  选择默认架构:\n\n")
		for i, opt := range archOptions {
			if i == m.archIdx {
				b.WriteString(fmt.Sprintf("  %s %s\n", theme.SelectedStyle.Render("▸"), theme.SelectedStyle.Render(opt)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", opt))
			}
		}
	case stepSaveDir:
		b.WriteString(fmt.Sprintf("  操作系统: %s  架构: %s\n\n", theme.HighlightStyle.Render(osOptions[m.osIdx]), theme.HighlightStyle.Render(archOptions[m.archIdx])))
		b.WriteString("  输入默认保存目录:\n\n")
		b.WriteString("  " + m.dirInput.View() + "\n")
	case stepConfirm:
		saveDir := m.dirInput.Value()
		if saveDir == "" {
			homeDir, _ := os.UserHomeDir()
			saveDir = filepath.Join(homeDir, "DockerImages")
		}
		b.WriteString("  确认配置:\n\n")
		b.WriteString(fmt.Sprintf("  操作系统:   %s\n", theme.HighlightStyle.Render(osOptions[m.osIdx])))
		b.WriteString(fmt.Sprintf("  架构:       %s\n", theme.HighlightStyle.Render(archOptions[m.archIdx])))
		b.WriteString(fmt.Sprintf("  保存目录:   %s\n", theme.HighlightStyle.Render(saveDir)))
		b.WriteString("\n  按 enter 保存配置")
	}

	if m.err != "" {
		b.WriteString("\n\n" + theme.ErrorStyle.Render("  "+m.err))
	}
	b.WriteString("\n\n" + theme.HelpStyle.Render("  ↑↓ 选择 · enter 确认"))
	return b.String()
}
