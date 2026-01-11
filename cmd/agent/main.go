package agent

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/orbit-w/orbit/config"
	"gitee.com/orbit-w/orbit/internal/agent"
	"gitee.com/orbit-w/orbit/lib/module/logger"
)

var (
	configPath = flag.String("config", "configs/config_center.yaml", "path to config file")
)

func main() {
	flag.Parse()

	err := config.InitConfig(*configPath)
	if err != nil {
		panic(err)
	}

	service, err := agent.Serve()
	if err != nil {
		panic(err)
	}

	gracefulShutdown(service)

	logger.GetLogger().Info("Agent server stopped successfully")
	logger.StopLogger()
}

// gracefulShutdown 优雅关闭服务
func gracefulShutdown(stopper func(ctx context.Context) error) {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT（Ctrl+C）和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 创建一个5分钟超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if stopper != nil {
		if err := stopper(ctx); err != nil {
			log.Printf("Error stopping stopper: %v", err)
		}
	}

	log.Println("Server exiting")
}
