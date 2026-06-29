package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Tool 是所有工具都需要实现的统一接口。
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string) error
}

// Registry 按名称保存当前可用的工具。
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Help() string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		fmt.Fprintf(&b, "  %-16s %s\n", name, r.tools[name].Description())
	}
	return b.String()
}
