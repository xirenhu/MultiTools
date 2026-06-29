package urlvisitor

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Bytes      int64
	Err        error
}

type TargetSnapshot struct {
	URL          string
	Total        int64
	Success      int64
	Errors       int64
	HTTPFailures int64
	Bytes        int64
}

type Snapshot struct {
	Total        int64
	Success      int64
	Errors       int64
	HTTPFailures int64
	StatusCodes  map[int]int64
	Bytes        int64
	AverageBytes int64
	Targets      []TargetSnapshot
	Average      time.Duration
	P95          time.Duration
	Max          time.Duration
}

type Metrics struct {
	mu           sync.Mutex
	total        int64
	success      int64
	errors       int64
	httpFailures int64
	statusCodes  map[int]int64
	latencies    []time.Duration
	bytes        int64
	targets      map[string]*TargetSnapshot
}

func NewMetrics() *Metrics {
	return &Metrics{
		statusCodes: make(map[int]int64),
		targets:     make(map[string]*TargetSnapshot),
	}
}

func (m *Metrics) Add(result Result) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.total++
	target := m.target(result.URL)
	target.Total++
	if result.Err != nil {
		m.errors++
		target.Errors++
		return
	}

	m.statusCodes[result.StatusCode]++
	if result.StatusCode >= 400 {
		m.httpFailures++
		target.HTTPFailures++
		return
	}

	m.success++
	target.Success++
	m.latencies = append(m.latencies, result.Latency)
	m.bytes += result.Bytes
	target.Bytes += result.Bytes
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	statusCodes := make(map[int]int64, len(m.statusCodes))
	for code, count := range m.statusCodes {
		statusCodes[code] = count
	}
	targets := make([]TargetSnapshot, 0, len(m.targets))
	for _, target := range m.targets {
		targets = append(targets, *target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].URL < targets[j].URL
	})

	latencies := append([]time.Duration(nil), m.latencies...)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var totalLatency time.Duration
	var max time.Duration
	for _, latency := range latencies {
		totalLatency += latency
		if latency > max {
			max = latency
		}
	}

	var avg time.Duration
	if len(latencies) > 0 {
		avg = totalLatency / time.Duration(len(latencies))
	}
	var averageBytes int64
	if m.success > 0 {
		averageBytes = m.bytes / m.success
	}

	return Snapshot{
		Total:        m.total,
		Success:      m.success,
		Errors:       m.errors,
		HTTPFailures: m.httpFailures,
		StatusCodes:  statusCodes,
		Bytes:        m.bytes,
		AverageBytes: averageBytes,
		Targets:      targets,
		Average:      avg,
		P95:          percentile(latencies, 95),
		Max:          max,
	}
}

func (m *Metrics) target(rawURL string) *TargetSnapshot {
	if rawURL == "" {
		rawURL = "-"
	}
	target, ok := m.targets[rawURL]
	if !ok {
		target = &TargetSnapshot{URL: rawURL}
		m.targets[rawURL] = target
	}
	return target
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := (len(sorted)*p + 99) / 100
	if index <= 0 {
		index = 1
	}
	if index > len(sorted) {
		index = len(sorted)
	}
	return sorted[index-1]
}

func (s Snapshot) StatusSummary() string {
	if len(s.StatusCodes) == 0 {
		return "-"
	}
	codes := make([]int, 0, len(s.StatusCodes))
	for code := range s.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, fmt.Sprintf("%d=%d", code, s.StatusCodes[code]))
	}
	return strings.Join(parts, ", ")
}

func FormatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)

	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/mb)
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/kb)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (s Snapshot) TargetSummary() string {
	if len(s.Targets) == 0 {
		return "-"
	}

	parts := make([]string, 0, len(s.Targets))
	for _, target := range s.Targets {
		parts = append(parts, fmt.Sprintf("%s 下载=%s 成功=%d HTTP失败=%d 错误=%d", target.URL, FormatBytes(target.Bytes), target.Success, target.HTTPFailures, target.Errors))
	}
	return strings.Join(parts, "；")
}
