package docker

import (
	"io"
	"net/http"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// ProgressRoundTripper 自定义 RoundTripper，用于更新进度条
type ProgressRoundTripper struct {
	rt  http.RoundTripper
	bar *progressbar.ProgressBar
}

// NewProgressRoundTripper 创建带进度条的 RoundTripper
func NewProgressRoundTripper(rt http.RoundTripper, bar *progressbar.ProgressBar) *ProgressRoundTripper {
	return &ProgressRoundTripper{
		rt:  rt,
		bar: bar,
	}
}

func (p *ProgressRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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
