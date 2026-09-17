package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"feetable/internal/feetable"
	"feetable/internal/httpapi"
	"feetable/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("feetable stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("用法：feetable init|serve|check|backup|restore --db <路径> [--out <备份路径>] [--from <恢复来源>]")
	}
	command := os.Args[1]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	dbPath := flags.String("db", env("FEETABLE_DB", "var/feetable.sqlite"), "SQLite 数据库路径")
	addr := flags.String("addr", env("FEETABLE_ADDR", "127.0.0.1:8081"), "HTTP 监听地址")
	output := flags.String("out", "", "备份目标（必须不存在）")
	source := flags.String("from", "", "恢复来源；--db 为新的目标路径")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("不接受额外参数")
	}
	ctx := context.Background()
	switch command {
	case "check":
		return feetable.CheckFile(ctx, *dbPath)
	case "restore":
		if *source == "" {
			return fmt.Errorf("restore 必须指定 --from")
		}
		return feetable.Restore(ctx, *source, *dbPath)
	case "init", "serve", "backup":
	default:
		return fmt.Errorf("未知命令 %s", command)
	}
	store, err := feetable.Open(*dbPath, command == "init")
	if err != nil {
		return err
	}
	defer store.Close()
	if command == "backup" {
		if *output == "" {
			return fmt.Errorf("backup 必须指定 --out")
		}
		return store.Backup(ctx, *output)
	}
	if command == "init" {
		slog.Info("数据库已创建", "path", *dbPath)
		return nil
	}
	server := &http.Server{Addr: *addr, Handler: httpapi.New(store, web.Assets()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 90 * time.Second}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { slog.Info("FeeTable 已启动", "addr", *addr); done <- server.ListenAndServe() }()
	select {
	case err = <-done:
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
