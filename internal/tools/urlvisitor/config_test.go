package urlvisitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigParsesDurations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
url: "https://example.com/"
duration: "1m"
timeout: "10s"
safety:
  require_authorization_confirm: false
  max_rate_per_second: 10
  allowed_hosts:
    - "example.com"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Duration != time.Minute {
		t.Fatalf("duration = %s，期望 %s", cfg.Duration, time.Minute)
	}
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("timeout = %s，期望 %s", cfg.Timeout, 10*time.Second)
	}
}

func TestLoadConfigSupportsURLs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
urls:
  - "https://example.com/"
  - "https://example.com/health"
strategy: "round_robin"
safety:
  require_authorization_confirm: false
  max_rate_per_second: 10
  allowed_hosts:
    - "example.com"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	targets := cfg.TargetURLs()
	if len(targets) != 2 {
		t.Fatalf("目标数量 = %d，期望 2", len(targets))
	}
}

func TestValidateRejectsURLAndURLsTogether(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com/"
	cfg.URLs = []string{"https://example.com/health"}

	if err := cfg.Validate(); err == nil {
		t.Fatal("期望同时配置 url 和 urls 时返回错误")
	}
}

func TestRoundRobinSelector(t *testing.T) {
	selector := NewTargetSelector("round_robin", []Target{
		{URL: "https://example.com/a"},
		{URL: "https://example.com/b"},
	})

	got := []string{
		selector.Next().URL,
		selector.Next().URL,
		selector.Next().URL,
	}
	want := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/a",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 次选择 = %s，期望 %s", i+1, got[i], want[i])
		}
	}
}

func TestFormatBytes(t *testing.T) {
	cases := map[int64]string{
		512:             "512 B",
		2048:            "2.00 KB",
		3 * 1024 * 1024: "3.00 MB",
	}

	for bytes, want := range cases {
		if got := FormatBytes(bytes); got != want {
			t.Fatalf("FormatBytes(%d) = %s，期望 %s", bytes, got, want)
		}
	}
}

func TestMetricsTracksBytesByTarget(t *testing.T) {
	metrics := NewMetrics()
	metrics.Add(Result{URL: "https://example.com/a.jpg", StatusCode: 200, Bytes: 70 * 1024})
	metrics.Add(Result{URL: "https://example.com/b.jpg", StatusCode: 200, Bytes: 10 * 1024})

	snapshot := metrics.Snapshot()
	if snapshot.Bytes != 80*1024 {
		t.Fatalf("总下载大小 = %s，期望 80.00 KB", FormatBytes(snapshot.Bytes))
	}
	if len(snapshot.Targets) != 2 {
		t.Fatalf("目标统计数量 = %d，期望 2", len(snapshot.Targets))
	}
}

func TestHostAllowed(t *testing.T) {
	if !hostAllowed("Example.com.", []string{"example.com"}) {
		t.Fatal("期望主机名忽略大小写后允许通过")
	}
	if hostAllowed("other.example.com", []string{"example.com"}) {
		t.Fatal("期望不匹配的主机名被拒绝")
	}
}
