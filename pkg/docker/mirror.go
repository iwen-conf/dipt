package docker

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"dipt/internal/logger"
	"dipt/internal/color"
	
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Mirror 镜像源信息
type Mirror struct {
	URL        string
	Available  bool
	Latency    time.Duration
	LastCheck  time.Time
	Priority   int  // 优先级，数字越小优先级越高
}

// MirrorManager 镜像源管理器
type MirrorManager struct {
	mirrors    []Mirror
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewMirrorManager 创建镜像源管理器
func NewMirrorManager(mirrorURLs []string) *MirrorManager {
	mirrors := make([]Mirror, len(mirrorURLs))
	for i, url := range mirrorURLs {
		mirrors[i] = Mirror{
			URL:       url,
			Available: true,
			Priority:  i,
		}
	}
	
	return &MirrorManager{
		mirrors: mirrors,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// AddMirror 添加镜像源
func (m *MirrorManager) AddMirror(url string, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 检查是否已存在
	for i := range m.mirrors {
		if m.mirrors[i].URL == url {
			m.mirrors[i].Priority = priority
			return
		}
	}
	
	m.mirrors = append(m.mirrors, Mirror{
		URL:       url,
		Available: true,
		Priority:  priority,
	})
}

// TestMirror 测试单个镜像源
func (m *MirrorManager) TestMirror(mirrorURL string) (bool, time.Duration, error) {
	// 构建测试 URL
	testURL := mirrorURL
	if !strings.HasPrefix(testURL, "http://") && !strings.HasPrefix(testURL, "https://") {
		testURL = "https://" + testURL
	}
	testURL = strings.TrimSuffix(testURL, "/") + "/v2/"
	
	start := time.Now()
	resp, err := m.httpClient.Get(testURL)
	latency := time.Since(start)
	
	if err != nil {
		return false, latency, err
	}
	defer resp.Body.Close()
	
	// 200 OK 或 401 Unauthorized 都表示镜像源可用
	// 401 表示需要认证，但服务是可达的
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
		return true, latency, nil
	}
	
	return false, latency, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

// CheckAllMirrors 检查所有镜像源的可用性
func (m *MirrorManager) CheckAllMirrors() {
	log := logger.GetLogger()
	log.Stage("检测镜像源可用性")
	
	var wg sync.WaitGroup
	results := make(chan struct {
		index     int
		available bool
		latency   time.Duration
		err       error
	}, len(m.mirrors))
	
	// 并发检测所有镜像源
	for i := range m.mirrors {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mirror := m.mirrors[idx]
			available, latency, err := m.TestMirror(mirror.URL)
			results <- struct {
				index     int
				available bool
				latency   time.Duration
				err       error
			}{idx, available, latency, err}
		}(i)
	}
	
	// 等待所有检测完成
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// 收集结果
	for result := range results {
		m.mu.Lock()
		m.mirrors[result.index].Available = result.available
		m.mirrors[result.index].Latency = result.latency
		m.mirrors[result.index].LastCheck = time.Now()
		m.mu.Unlock()
		
		if result.available {
			log.Success("镜像源 %s 可用 (延迟: %v)", 
				m.mirrors[result.index].URL, result.latency)
		} else {
			log.Warning("镜像源 %s 不可用: %v", 
				m.mirrors[result.index].URL, result.err)
		}
	}
}

// GetAvailableMirrors 获取可用的镜像源（按优先级和延迟排序）
func (m *MirrorManager) GetAvailableMirrors() []Mirror {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var available []Mirror
	for _, mirror := range m.mirrors {
		if mirror.Available {
			available = append(available, mirror)
		}
	}
	
	// 按优先级和延迟排序
	// 优先级高的在前，同优先级按延迟排序
	for i := 0; i < len(available)-1; i++ {
		for j := i + 1; j < len(available); j++ {
			if available[i].Priority > available[j].Priority ||
				(available[i].Priority == available[j].Priority && 
				 available[i].Latency > available[j].Latency) {
				available[i], available[j] = available[j], available[i]
			}
		}
	}
	
	return available
}

// TryPullWithMirrors 尝试使用镜像源拉取镜像
func (m *MirrorManager) TryPullWithMirrors(
	ref name.Reference, 
	options []remote.Option, 
	callback func(mirrorRef name.Reference, mirrorURL string) error,
) error {
	log := logger.GetLogger()
	
	// 先检查所有镜像源
	m.CheckAllMirrors()
	
	// 获取可用的镜像源
	availableMirrors := m.GetAvailableMirrors()
	
	if len(availableMirrors) == 0 {
		log.Warning("没有可用的镜像源，将使用原始地址")
		return callback(ref, "")
	}
	
	var lastErr error
	
	// 尝试每个可用的镜像源
	for _, mirror := range availableMirrors {
		log.Info("尝试使用镜像源: %s", mirror.URL)
		
		// 创建镜像引用
		mirrorRef, err := CreateMirrorReference(ref, mirror.URL)
		if err != nil {
			log.Warning("创建镜像引用失败: %v", err)
			continue
		}
		
		// 尝试拉取
		err = callback(mirrorRef, mirror.URL)
		if err == nil {
			// 成功
			color.Success("成功使用镜像源: %s", mirror.URL)
			return nil
		}
		
		lastErr = err
		log.Warning("镜像源 %s 拉取失败: %v", mirror.URL, err)
		
		// 标记该镜像源暂时不可用
		m.mu.Lock()
		for i := range m.mirrors {
			if m.mirrors[i].URL == mirror.URL {
				m.mirrors[i].Available = false
				break
			}
		}
		m.mu.Unlock()
	}
	
	// 所有镜像源都失败了，尝试使用原始地址
	log.Warning("所有镜像源都失败，尝试使用原始地址")
	err := callback(ref, "")
	if err != nil {
		return fmt.Errorf("所有镜像源和原始地址都失败: %w", lastErr)
	}
	
	return nil
}

// CreateMirrorReference 创建镜像源引用
func CreateMirrorReference(ref name.Reference, mirrorURL string) (name.Reference, error) {
	// 移除协议前缀
	mirrorURL = strings.TrimPrefix(mirrorURL, "http://")
	mirrorURL = strings.TrimPrefix(mirrorURL, "https://")
	mirrorURL = strings.TrimSuffix(mirrorURL, "/")
	
	// 获取原始镜像名称
	originalName := ref.Context().RepositoryStr()
	id := ref.Identifier()
	
	// 构建新的引用
	var newRef string
	if strings.Contains(id, ":") && strings.HasPrefix(id, "sha256:") {
		// Digest 形式
		newRef = fmt.Sprintf("%s/%s@%s", mirrorURL, originalName, id)
	} else {
		// Tag 形式
		newRef = fmt.Sprintf("%s/%s:%s", mirrorURL, originalName, id)
	}
	
	return name.ParseReference(newRef)
}

// IsDockerHubImage 判断是否是 Docker Hub 镜像
func IsDockerHubImage(ref name.Reference) bool {
	registry := ref.Context().Registry.Name()
	return registry == "docker.io" || 
		   registry == "registry-1.docker.io" || 
		   registry == "index.docker.io"
}
