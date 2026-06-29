package urlvisitor

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"
)

type Tool struct{}

func New() *Tool {
	return &Tool{}
}

func (t *Tool) Name() string {
	return "url-visitor"
}

func (t *Tool) Description() string {
	return "已授权 URL 持续访问工具，支持限速和指标统计"
}

func (t *Tool) Run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "", "YAML 配置文件路径")
	fs.Usage = func() {
		fmt.Println("用法：")
		fmt.Println("  multitools url-visitor --config <配置文件路径>")
		fmt.Println()
		fmt.Println("参数：")
		fmt.Println("  --config    YAML 配置文件路径")
	}
	if wantsHelp(args) {
		fs.Usage()
		return nil
	}
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("参数解析失败，请检查命令参数")
	}
	if *configPath == "" {
		return fmt.Errorf("缺少必填参数：--config")
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		return err
	}

	fmt.Printf("启动 url-visitor：目标数=%d 方法=%s 速率=%d次/秒 并发=%d 策略=%s 持续时间=%s\n",
		len(cfg.TargetURLs()),
		cfg.Method,
		cfg.RatePerSecond,
		cfg.Concurrency,
		cfg.Strategy,
		formatDuration(cfg.Duration),
	)

	return NewRunner(cfg).Run(ctx)
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func formatDuration(duration time.Duration) string {
	if duration == 0 {
		return "直到手动停止"
	}
	return duration.String()
}
