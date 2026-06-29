package app

import (
	"context"
	"fmt"
	"strings"

	"multitools/internal/tools"
)

// Run 根据命令行参数分发到对应工具。应用层保持轻量，让各个工具可以独立演进。
func Run(ctx context.Context, registry *tools.Registry, args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printUsage(registry)
		return nil
	}

	name := args[0]
	tool, ok := registry.Get(name)
	if !ok {
		return fmt.Errorf("未知工具 %q\n\n可用工具：\n%s", name, registry.Help())
	}

	return tool.Run(ctx, args[1:])
}

func printUsage(registry *tools.Registry) {
	fmt.Println("MultiTools")
	fmt.Println()
	fmt.Println("用法：")
	fmt.Println("  multitools <tool> [options]")
	fmt.Println()
	fmt.Println("可用工具：")
	fmt.Println(strings.TrimRight(registry.Help(), "\n"))
}
