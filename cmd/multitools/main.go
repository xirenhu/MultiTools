package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"multitools/internal/app"
	"multitools/internal/tools"
	"multitools/internal/tools/urlvisitor"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	registry := tools.NewRegistry()
	registry.Register(urlvisitor.New())

	if err := app.Run(ctx, registry, os.Args[1:]); err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		}
		os.Exit(1)
	}
}
