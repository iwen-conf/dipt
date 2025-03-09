package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/schollz/progressbar/v3"
)

// Config 定义 JSON 配置文件的结构
type Config struct {
	Registry struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"registry"`
}

// Platform 定义平台信息
type Platform struct {
	OS   string
	Arch string
}

// loadConfig 读取和解析 config.json 文件
func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return config, nil
}

// progressRoundTripper 自定义 RoundTripper，用于更新进度条
type progressRoundTripper struct {
	rt  http.RoundTripper
	bar *progressbar.ProgressBar
}

func (p *progressRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := p.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if strings.Contains(req.URL.Path, "/blobs/") {
		resp.Body = &progressReader{reader: resp.Body, bar: p.bar, closer: resp.Body}
	}
	return resp, nil
}

// progressReader 包装 io.Reader，读取数据时更新进度条
type progressReader struct {
	reader io.Reader
	bar    *progressbar.ProgressBar
	closer io.Closer // 保存原始的 io.Closer
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.bar.Add(n)
	}
	return
}

func (pr *progressReader) Close() error {
	if pr.closer != nil {
		return pr.closer.Close()
	}
	return nil
}

// pullAndSaveImage 拉取镜像并保存为 tar 文件，带进度显示
func pullAndSaveImage(imageName, outputFile string, platform Platform, config Config) error {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return fmt.Errorf("解析镜像名称失败: %v", err)
	}

	var auth authn.Authenticator
	if config.Registry.Username != "" && config.Registry.Password != "" {
		auth = authn.FromConfig(authn.AuthConfig{
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		})
	} else {
		auth = authn.Anonymous
	}

	options := []remote.Option{remote.WithAuth(auth)}

	// 添加平台选项
	options = append(options, remote.WithPlatform(v1.Platform{
		OS:           platform.OS,
		Architecture: platform.Arch,
	}))

	desc, err := remote.Get(ref, options...)
	if err != nil {
		return fmt.Errorf("获取镜像描述失败: %v", err)
	}

	img, err := desc.Image()
	if err != nil {
		return fmt.Errorf("获取 v1.Image 失败: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("获取层信息失败: %v", err)
	}

	var totalSize int64
	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return fmt.Errorf("获取层大小失败: %v", err)
		}
		totalSize += size
	}

	bar := progressbar.NewOptions64(totalSize,
		progressbar.OptionSetDescription("拉取镜像中"),
		progressbar.OptionShowBytes(true),
	)

	rt := &progressRoundTripper{rt: http.DefaultTransport, bar: bar}
	options = append(options, remote.WithTransport(rt))
	img, err = remote.Image(ref, options...)
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}

	tarFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建 tar 文件失败: %v", err)
	}
	defer tarFile.Close()

	err = tarball.WriteToFile(outputFile, ref, img)
	if err != nil {
		return fmt.Errorf("保存镜像到 tar 文件失败: %v", err)
	}

	return nil
}

// 从镜像名称中提取软件名和版本
func parseImageName(imageName string) (software, version string) {
	parts := strings.Split(imageName, ":")
	if len(parts) < 2 {
		return parts[0], "latest"
	}

	// 处理软件名称中可能包含的路径
	nameParts := strings.Split(parts[0], "/")
	software = nameParts[len(nameParts)-1]

	return software, parts[1]
}

func generateOutputFileName(imageName string, platform Platform) string {
	software, version := parseImageName(imageName)
	return fmt.Sprintf("%s_%s_%s_%s.tar", software, version, platform.OS, platform.Arch)
}

func parseArgs() (imageName string, outputFile string, platform Platform, err error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return "", "", Platform{}, fmt.Errorf("用法: dipt [-os <系统>] [-arch <架构>] <镜像名称> [输出文件]")
	}

	// 设置默认值
	platform = Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-os":
			if i+1 >= len(args) {
				return "", "", Platform{}, fmt.Errorf("-os 参数需要指定系统名称")
			}
			platform.OS = args[i+1]
			i++
		case "-arch":
			if i+1 >= len(args) {
				return "", "", Platform{}, fmt.Errorf("-arch 参数需要指定架构名称")
			}
			platform.Arch = args[i+1]
			i++
		default:
			// 如果不是选项参数，则认为是镜像名称或输出文件
			if imageName == "" {
				imageName = args[i]
			} else {
				outputFile = args[i]
			}
		}
	}

	if imageName == "" {
		return "", "", Platform{}, fmt.Errorf("必须指定镜像名称")
	}

	// 如果没有指定输出文件，则根据镜像信息生成
	if outputFile == "" {
		outputFile = generateOutputFileName(imageName, platform)
	}

	return imageName, outputFile, platform, nil
}

func main() {
	imageName, outputFile, platform, err := parseArgs()
	if err != nil {
		fmt.Println("错误:", err)
		fmt.Println("示例: dipt -os linux -arch amd64 nginx:latest [output.tar]")
		os.Exit(1)
	}

	// 加载配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	// 拉取镜像并保存
	fmt.Printf("正在拉取镜像 %s (系统: %s, 架构: %s)...\n", imageName, platform.OS, platform.Arch)
	err = pullAndSaveImage(imageName, outputFile, platform, config)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	fmt.Printf("\n镜像已保存到 %s\n", outputFile)
}
