package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"xiaozhi-server-go/internal/workflow"
)

func main() {
	log.Println("Starting XiaoZhi Flow Workflow Engine Example...")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 运行示例
	go func() {
		workflow.RunExample()
	}()

	// 等待信号
	<-sigChan
	log.Println("\nReceived shutdown signal, exiting...")
}