package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/schollz/progressbar/v3"
)

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
