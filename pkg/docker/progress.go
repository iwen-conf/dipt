package docker

import (
	"io"
	"net/http"
	"strings"
)

// ProgressCallback 进度回调函数
type ProgressCallback func(downloaded, total int64)

// CallbackRoundTripper 自定义 RoundTripper，通过回调报告进度
type CallbackRoundTripper struct {
	rt       http.RoundTripper
	callback ProgressCallback
}

// NewCallbackRoundTripper 创建带回调的 RoundTripper
func NewCallbackRoundTripper(rt http.RoundTripper, cb ProgressCallback) *CallbackRoundTripper {
	return &CallbackRoundTripper{
		rt:       rt,
		callback: cb,
	}
}

func (c *CallbackRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if c.callback != nil && strings.Contains(req.URL.Path, "/blobs/") {
		resp.Body = &callbackReader{
			reader:   resp.Body,
			callback: c.callback,
			closer:   resp.Body,
		}
	}
	return resp, nil
}

// callbackReader 包装 io.Reader，读取数据时调用回调
type callbackReader struct {
	reader     io.Reader
	callback   ProgressCallback
	closer     io.Closer
	downloaded int64
}

func (cr *callbackReader) Read(p []byte) (n int, err error) {
	n, err = cr.reader.Read(p)
	if n > 0 {
		cr.downloaded += int64(n)
		if cr.callback != nil {
			cr.callback(cr.downloaded, 0) // total 由外层设置
		}
	}
	return
}

func (cr *callbackReader) Close() error {
	if cr.closer != nil {
		return cr.closer.Close()
	}
	return nil
}

// TotalTrackingRoundTripper 带总量追踪的 RoundTripper
type TotalTrackingRoundTripper struct {
	rt         http.RoundTripper
	callback   ProgressCallback
	totalSize  int64
	downloaded int64
}

// NewTotalTrackingRoundTripper 创建带总量追踪的 RoundTripper
func NewTotalTrackingRoundTripper(rt http.RoundTripper, totalSize int64, cb ProgressCallback) *TotalTrackingRoundTripper {
	return &TotalTrackingRoundTripper{
		rt:        rt,
		callback:  cb,
		totalSize: totalSize,
	}
}

func (t *TotalTrackingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if t.callback != nil && strings.Contains(req.URL.Path, "/blobs/") {
		resp.Body = &totalTrackingReader{
			reader:   resp.Body,
			closer:   resp.Body,
			parent:   t,
		}
	}
	return resp, nil
}

type totalTrackingReader struct {
	reader io.Reader
	closer io.Closer
	parent *TotalTrackingRoundTripper
}

func (r *totalTrackingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.parent.downloaded += int64(n)
		if r.parent.callback != nil {
			r.parent.callback(r.parent.downloaded, r.parent.totalSize)
		}
	}
	return
}

func (r *totalTrackingReader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}
