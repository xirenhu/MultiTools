package urlvisitor

import (
	"context"
	"io"
	"net/http"
	"time"
)

func NewHTTPClient(cfg Config) *http.Client {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	if !cfg.FollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}

func Visit(ctx context.Context, client *http.Client, cfg Config, target Target) Result {
	req, err := http.NewRequestWithContext(ctx, cfg.Method, target.URL, nil)
	if err != nil {
		return Result{URL: target.URL, Err: err}
	}

	req.Header.Set("User-Agent", pickUserAgent(cfg.UserAgents))
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Connection", "keep-alive")
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: target.URL, Err: err}
	}
	defer resp.Body.Close()

	bytesRead, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return Result{URL: target.URL, Err: err}
	}

	return Result{
		URL:        target.URL,
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		Bytes:      bytesRead,
	}
}
