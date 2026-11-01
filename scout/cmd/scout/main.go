// Command scout 是給 AI agent 操作的科技資訊雷達 CLI：
// 抓取你自選的來源、去重、用可解釋的規則排序，挑選與摘要交給 agent。
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Andy880828/personal-lab/scout/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
