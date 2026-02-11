package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"dipt/internal/config"
	"dipt/internal/tui/components"
	"dipt/internal/tui/theme"
	"dipt/internal/types"
	"dipt/pkg/docker"

	tea "github.com/charmbracelet/bubbletea"
)

// AppState 应用状态
type AppState int

const (
	StateSetup    AppState = iota // 首次运行配置向导
	StateMenu                    // 主菜单
	StatePullForm                // 拉取表单
	StatePulling                 // 拉取进度
	StateSettings                // 设置
	StateMirrors                 // 镜像源管理
)

// programRef 共享引用，解决 Bubble Tea 值拷贝导致 program 为 nil 的问题
type programRef struct {
	p *tea.Program
}

// AppModel 顶层应用模型
type AppModel struct {
	state      AppState
	userConfig *types.UserConfig
	effConfig  types.Config
	width      int
	height     int

	// 子模型
	setup    components.SetupModel
	menu     components.MenuModel
	pullForm components.PullFormModel
	pullProg components.PullProgressModel
	settings components.SettingsModel
	mirrors  components.MirrorsModel

	// tea.Program 共享引用，所有副本共享同一个指针
	program *programRef
}

// NewApp 创建应用
func NewApp() AppModel {
	userCfg, effCfg, err := config.LoadEffectiveConfigs()
	if err != nil || userCfg == nil {
		// 需要首次配置
		return AppModel{
			state:   StateSetup,
			setup:   components.NewSetupModel(),
			program: &programRef{},
		}
	}

	return AppModel{
		state:      StateMenu,
		userConfig: userCfg,
		effConfig:  effCfg,
		menu:       components.NewMenuModel(),
		program:    &programRef{},
	}
}

func (m AppModel) Init() tea.Cmd {
	switch m.state {
	case StateSetup:
		return m.setup.Init()
	case StateMenu:
		return m.menu.Init()
	case StatePulling:
		return m.pullProg.Init()
	default:
		return nil
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 全局消息处理
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// 仅在菜单状态下 q 退出
		if msg.String() == "q" && m.state == StateMenu {
			return m, tea.Quit
		}
	}

	// 路由到子模型
	switch m.state {
	case StateSetup:
		return m.updateSetup(msg)
	case StateMenu:
		return m.updateMenu(msg)
	case StatePullForm:
		return m.updatePullForm(msg)
	case StatePulling:
		return m.updatePulling(msg)
	case StateSettings:
		return m.updateSettings(msg)
	case StateMirrors:
		return m.updateMirrors(msg)
	}
	return m, nil
}

func (m AppModel) View() string {
	var content string
	switch m.state {
	case StateSetup:
		content = m.setup.View()
	case StateMenu:
		content = m.menu.View()
	case StatePullForm:
		content = m.pullForm.View()
	case StatePulling:
		content = m.pullProg.View()
	case StateSettings:
		content = m.settings.View()
	case StateMirrors:
		content = m.mirrors.View()
	}
	return theme.AppStyle.Render(content)
}

// --- 子模型更新 ---

func (m AppModel) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.SetupDoneMsg:
		m.userConfig = msg.Config
		_, effCfg, _ := config.LoadEffectiveConfigs()
		m.effConfig = effCfg
		m.state = StateMenu
		m.menu = components.NewMenuModel()
		return m, m.menu.Init()
	}
	var cmd tea.Cmd
	m.setup, cmd = m.setup.Update(msg)
	return m, cmd
}

func (m AppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.MenuChosenMsg:
		switch msg.Choice {
		case components.MenuPull:
			m.state = StatePullForm
			m.pullForm = components.NewPullFormModel(m.userConfig)
			return m, m.pullForm.Init()
		case components.MenuSettings:
			m.state = StateSettings
			m.settings = components.NewSettingsModel(m.userConfig)
			return m, m.settings.Init()
		case components.MenuMirrors:
			m.state = StateMirrors
			m.mirrors = components.NewMirrorsModel(m.userConfig)
			return m, m.mirrors.Init()
		case components.MenuQuit:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m AppModel) updatePullForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.BackToMenuMsg:
		m.state = StateMenu
		m.menu = components.NewMenuModel()
		return m, m.menu.Init()
	case components.StartPullMsg:
		m.state = StatePulling
		m.pullProg = components.NewPullProgressModel(msg.ImageName)

		// 计算输出文件
		outputFile := msg.OutputFile
		if outputFile == "" {
			outputFile = docker.GenerateOutputFileName(msg.ImageName, msg.Platform)
			if m.userConfig != nil && m.userConfig.DefaultSaveDir != "" {
				outputFile = filepath.Join(m.userConfig.DefaultSaveDir, outputFile)
			}
		}
		_ = os.MkdirAll(filepath.Dir(outputFile), 0755)

		// 重新加载配置以获取最新镜像源
		_, effCfg, _ := config.LoadEffectiveConfigs()
		m.effConfig = effCfg

		// 启动异步拉取
		return m, tea.Batch(
			m.pullProg.Init(),
			m.startPull(msg.ImageName, outputFile, msg.Platform),
		)
	}
	var cmd tea.Cmd
	m.pullForm, cmd = m.pullForm.Update(msg)
	return m, cmd
}

func (m AppModel) updatePulling(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.BackToMenuMsg:
		m.state = StateMenu
		m.menu = components.NewMenuModel()
		return m, m.menu.Init()
	default:
		_ = msg
	}
	var cmd tea.Cmd
	m.pullProg, cmd = m.pullProg.Update(msg)
	return m, cmd
}

func (m AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case components.BackToMenuMsg:
		// 重新加载配置
		userCfg, effCfg, _ := config.LoadEffectiveConfigs()
		if userCfg != nil {
			m.userConfig = userCfg
		}
		m.effConfig = effCfg
		m.state = StateMenu
		m.menu = components.NewMenuModel()
		return m, m.menu.Init()
	}
	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

func (m AppModel) updateMirrors(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case components.BackToMenuMsg:
		// 重新加载配置
		userCfg, effCfg, _ := config.LoadEffectiveConfigs()
		if userCfg != nil {
			m.userConfig = userCfg
		}
		m.effConfig = effCfg
		m.state = StateMenu
		m.menu = components.NewMenuModel()
		return m, m.menu.Init()
	}
	var cmd tea.Cmd
	m.mirrors, cmd = m.mirrors.Update(msg)
	return m, cmd
}

// startPull 启动异步拉取
func (m AppModel) startPull(imageName, outputFile string, platform types.Platform) tea.Cmd {
	return func() tea.Msg {
		opts := docker.PullOptions{
			ImageName:  imageName,
			OutputFile: outputFile,
			Platform:   platform,
			Config:     m.effConfig,
			OnProgress: func(downloaded, total int64) {
				if m.program.p != nil {
					m.program.p.Send(components.ProgressMsg{
						Downloaded: downloaded,
						Total:      total,
					})
				}
			},
			OnLog: func(level, msg string) {
				if m.program.p != nil {
					m.program.p.Send(components.LogMsg{
						Level:   level,
						Message: msg,
					})
				}
			},
		}

		err := docker.PullAndSave(opts)
		return components.PullDoneMsg{Err: err}
	}
}

// Run 启动 TUI 应用
func Run() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	app.program.p = p

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI 运行失败: %w", err)
	}
	_ = finalModel
	return nil
}
