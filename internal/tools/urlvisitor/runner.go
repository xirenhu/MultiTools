package urlvisitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Runner struct {
	cfg      Config
	targets  []Target
	selector TargetSelector
	metrics  *Metrics
}

func NewRunner(cfg Config) *Runner {
	targets := make([]Target, 0, len(cfg.TargetURLs()))
	for _, targetURL := range cfg.TargetURLs() {
		targets = append(targets, Target{URL: targetURL})
	}

	return &Runner{
		cfg:      cfg,
		targets:  targets,
		selector: NewTargetSelector(cfg.Strategy, targets),
		metrics:  NewMetrics(),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if err := SafetyCheck(ctx, r.cfg); err != nil {
		return err
	}

	runCtx := ctx
	cancel := func() {}
	if r.cfg.Duration > 0 {
		runCtx, cancel = context.WithTimeout(ctx, r.cfg.Duration)
	}
	defer cancel()

	client := NewHTTPClient(r.cfg)
	jobs := make(chan Target)
	results := make(chan Result, r.cfg.Concurrency)

	var workers sync.WaitGroup
	for i := 0; i < r.cfg.Concurrency; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for target := range jobs {
				result := Visit(runCtx, client, r.cfg, target)
				if errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, context.DeadlineExceeded) {
					continue
				}
				results <- result
			}
		}()
	}

	var collectors sync.WaitGroup
	collectors.Add(1)
	go func() {
		defer collectors.Done()
		for result := range results {
			r.metrics.Add(result)
		}
	}()

	reportDone := make(chan struct{})
	go r.report(runCtx, reportDone)

	err := r.produce(runCtx, jobs)
	close(jobs)
	workers.Wait()
	close(results)
	collectors.Wait()
	<-reportDone

	r.printSnapshot("最终统计")
	return err
}

func (r *Runner) produce(ctx context.Context, jobs chan<- Target) error {
	interval := time.Second / time.Duration(r.cfg.RatePerSecond)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			select {
			case jobs <- r.selector.Next():
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (r *Runner) report(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.printSnapshot("运行中")
		}
	}
}

func (r *Runner) printSnapshot(label string) {
	s := r.metrics.Snapshot()
	fmt.Printf(
		"[%s] 目标数=%d 总数=%d 成功=%d HTTP失败=%d 错误=%d 下载=%s 平均每次下载=%s 状态码=%s 平均耗时=%s P95=%s 最大耗时=%s\n",
		label,
		len(r.targets),
		s.Total,
		s.Success,
		s.HTTPFailures,
		s.Errors,
		FormatBytes(s.Bytes),
		FormatBytes(s.AverageBytes),
		s.StatusSummary(),
		s.Average.Round(time.Millisecond),
		s.P95.Round(time.Millisecond),
		s.Max.Round(time.Millisecond),
	)
	fmt.Printf("  按目标统计：%s\n", s.TargetSummary())
}
