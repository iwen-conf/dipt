package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "golang.org/x/term"
)

// UserConfig 用户配置结构
type UserConfig struct {
	DefaultOS      string   `json:"default_os"`       // 默认操作系统
	DefaultArch    string   `json:"default_arch"`     // 默认架构
	DefaultSaveDir string   `json:"default_save_dir"` // 默认保存目录
	Registry       Registry `json:"registry"`         // 镜像仓库配置
}

// Registry 镜像仓库配置
type Registry struct {
	Mirrors  []string `json:"mirrors,omitempty"`  // 镜像加速器列表
	Username string   `json:"username,omitempty"` // 用户名
	Password string   `json:"password,omitempty"` // 密码
}

const configFileName = ".dipt_config"

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %v", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// loadUserConfig 加载用户配置
func loadUserConfig() (*UserConfig, error) {
    configPath, err := getConfigFilePath()
    if err != nil {
        return nil, err
    }

    // 如果配置文件不存在，启动交互式配置
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        // 支持在非交互环境或设置了禁用交互时，使用默认配置
        noInteractive := os.Getenv("DIPT_NO_INTERACTIVE") == "1"
        isTTY := term.IsTerminal(int(os.Stdin.Fd()))
        if noInteractive || !isTTY {
            // 从环境变量读取默认值，否则使用内置默认
            defOS := os.Getenv("DIPT_DEFAULT_OS")
            if defOS == "" {
                defOS = "linux"
            }
            defArch := os.Getenv("DIPT_DEFAULT_ARCH")
            if defArch == "" {
                defArch = "amd64"
            }
            homeDir, _ := os.UserHomeDir()
            defSave := os.Getenv("DIPT_DEFAULT_SAVE_DIR")
            if defSave == "" {
                defSave = filepath.Join(homeDir, "DockerImages")
            }

            // 确保目录存在
            _ = os.MkdirAll(defSave, 0755)

            cfg := &UserConfig{
                DefaultOS:      defOS,
                DefaultArch:    defArch,
                DefaultSaveDir: defSave,
            }
            // 尝试落盘但不阻塞
            _ = saveUserConfig(cfg)
            return cfg, nil
        }
        fmt.Printf("未检测到配置文件 %s\n", configPath)
        return interactiveConfig()
    }

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// saveUserConfig 保存用户配置
func saveUserConfig(config *UserConfig) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// setConfigValue 设置配置值
func setConfigValue(key, value string) error {
	config, err := loadUserConfig()
	if err != nil {
		return err
	}

	switch key {
	case "os":
		if !isValidOS(value) {
			return fmt.Errorf("不支持的操作系统: %s", value)
		}
		config.DefaultOS = value
	case "arch":
		if !isValidArch(value) {
			return fmt.Errorf("不支持的架构: %s", value)
		}
		config.DefaultArch = value
	case "save_dir":
		// 验证目录是否存在
		if _, err := os.Stat(value); os.IsNotExist(err) {
			// 尝试创建目录
			err = os.MkdirAll(value, 0755)
			if err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		}
		// 转换为绝对路径
		absPath, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("转换路径失败: %v", err)
		}
		config.DefaultSaveDir = absPath
	case "mirror":
		return fmt.Errorf("请使用 mirror 相关的子命令管理镜像加速器:\n" +
			"  dipt mirror list          # 列出所有镜像加速器\n" +
			"  dipt mirror add <URL>     # 添加镜像加速器\n" +
			"  dipt mirror del <URL>     # 删除镜像加速器\n" +
			"  dipt mirror clear         # 清空所有镜像加速器")
	default:
		return fmt.Errorf("未知的配置项: %s", key)
	}

	return saveUserConfig(config)
}

// handleMirrorCommand 处理镜像加速器相关命令
func handleMirrorCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("缺少子命令，可用命令：list, add, del, clear, test")
	}

	config, err := loadUserConfig()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		if len(config.Registry.Mirrors) == 0 {
			fmt.Println("当前未配置任何镜像加速器")
			return nil
		}
		fmt.Println("已配置的镜像加速器：")
		for i, mirror := range config.Registry.Mirrors {
			fmt.Printf("%d. %s\n", i+1, mirror)
		}

	case "add":
		if len(args) != 2 {
			return fmt.Errorf("用法: dipt mirror add <URL>")
		}
		mirror := args[1]
		// 检查是否已存在
		for _, m := range config.Registry.Mirrors {
			if m == mirror {
				return fmt.Errorf("镜像加速器已存在: %s", mirror)
			}
		}
		config.Registry.Mirrors = append(config.Registry.Mirrors, mirror)
		err = saveUserConfig(config)
		if err != nil {
			return err
		}
		fmt.Printf("✅ 已添加镜像加速器: %s\n", mirror)

	case "del":
		if len(args) != 2 {
			return fmt.Errorf("用法: dipt mirror del <URL>")
		}
		mirror := args[1]
		found := false
		newMirrors := make([]string, 0)
		for _, m := range config.Registry.Mirrors {
			if m != mirror {
				newMirrors = append(newMirrors, m)
			} else {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("未找到指定的镜像加速器: %s", mirror)
		}
		config.Registry.Mirrors = newMirrors
		err = saveUserConfig(config)
		if err != nil {
			return err
		}
		fmt.Printf("✅ 已删除镜像加速器: %s\n", mirror)

	case "clear":
		config.Registry.Mirrors = []string{}
		err = saveUserConfig(config)
		if err != nil {
			return err
		}
		fmt.Println("✅ 已清空所有镜像加速器")

    case "test":
        if len(args) != 2 {
            return fmt.Errorf("用法: dipt mirror test <URL>")
        }
        mirror := args[1]
        url := mirror
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
		if !strings.HasSuffix(url, "/v2/") {
			if strings.HasSuffix(url, "/") {
				url = url + "v2/"
			} else {
				url = url + "/v2/"
			}
		}
        fmt.Printf("正在测试镜像加速器连通性: %s ...\n", url)
        client := &http.Client{Timeout: 5 * time.Second}
        resp, err := client.Get(url)
        if err != nil {
            fmt.Printf("❌ 连接失败: %v\n", err)
            return nil
        }
        defer resp.Body.Close()
		fmt.Printf("返回状态码: %d\n", resp.StatusCode)
		fmt.Println("响应头:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		body := make([]byte, 512)
		n, _ := resp.Body.Read(body)
		if n > 0 {
			fmt.Println("响应体(前512字节):")
			fmt.Println(string(body[:n]))
		}
		if resp.StatusCode == 200 {
			fmt.Println("✅ 连接成功 (200)，该加速器可用")
		} else if resp.StatusCode == 401 {
			fmt.Println("✅ 连接成功 (401)，需要认证，通常也代表加速器可用")
		} else {
			fmt.Println("⚠️ 连接异常，状态码请参考上方信息")
		}
		return nil

	default:
		return fmt.Errorf("未知的子命令: %s", args[0])
	}

	return nil
}

// loadProjectConfig 加载项目级配置 ./config.json（如果存在）
func loadProjectConfig() (*Config, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, fmt.Errorf("读取项目配置失败: %v", err)
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("解析项目配置失败: %v", err)
    }
    return &cfg, nil
}

// effectiveRegistry 合并生效的 Registry（优先级：环境变量 > 项目配置 > 用户配置）
func effectiveRegistry(user *UserConfig, project *Config) Config {
    var out Config
    // 先复制用户配置
    if user != nil {
        out.Registry.Mirrors = append(out.Registry.Mirrors, user.Registry.Mirrors...)
        out.Registry.Username = user.Registry.Username
        out.Registry.Password = user.Registry.Password
    }
    // 项目配置覆盖
    if project != nil {
        if len(project.Registry.Mirrors) > 0 {
            out.Registry.Mirrors = project.Registry.Mirrors
        }
        if project.Registry.Username != "" {
            out.Registry.Username = project.Registry.Username
        }
        if project.Registry.Password != "" {
            out.Registry.Password = project.Registry.Password
        }
    }
    // 环境变量最终覆盖
    if u := os.Getenv("DIPT_REGISTRY_USERNAME"); u != "" {
        out.Registry.Username = u
    }
    if p := os.Getenv("DIPT_REGISTRY_PASSWORD"); p != "" {
        out.Registry.Password = p
    }
    if m := os.Getenv("DIPT_REGISTRY_MIRRORS"); m != "" {
        // 逗号分隔
        parts := strings.Split(m, ",")
        mirrors := make([]string, 0, len(parts))
        for _, s := range parts {
            s = strings.TrimSpace(s)
            if s != "" {
                mirrors = append(mirrors, s)
            }
        }
        if len(mirrors) > 0 {
            out.Registry.Mirrors = mirrors
        }
    }
    return out
}

// loadEffectiveConfigs 载入用户配置与项目配置，并返回合并后的 Registry 配置
func loadEffectiveConfigs() (*UserConfig, Config, error) {
    userCfg, err := loadUserConfig()
    if err != nil {
        return nil, Config{}, err
    }
    projCfg, err := loadProjectConfig()
    if err != nil {
        return nil, Config{}, err
    }
    eff := effectiveRegistry(userCfg, projCfg)
    return userCfg, eff, nil
}
