package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

func main() {
    run([]byte(`{
        "inbounds": [
            {
                "type": "shadowsocks",
                "tag": "shadowsocks-in",
                "listen": "127.0.0.1",
                "method": "2022-blake3-aes-128-gcm",
                "password": "YWVzLTEyOC1nY206aGFNTE1YaXJCeW42ckdWaA==",
                "multiplex": {
                    "enabled": true
                }
            }
        ],
        "outbounds": [
            {
                "type": "direct"
            }
        ]
    }`));
}

func run(config []byte) error {
	// 创建一个可以被取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 读取配置文件
	var opts option.Options
	var err = json.Unmarshal(config, &opts)
	if err != nil {
		return err
	}

	// 创建一个新的sing-box实例
	instance, err := box.New(box.Options{
		Context: ctx,
		Options: opts,
	})
	if err != nil {
		return err
	}

	// 设置操作系统信号处理
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(osSignals)

	go func() {
		<-osSignals
		cancel() // 当接收到信号时取消上下文
		if err := instance.Close(); err != nil {
			log.Error("Failed to close sing-box:", err)
		}
	}()

	// 启动sing-box实例
	if err := instance.Start(); err != nil {
		return err
	}

	// 等待上下文被取消，表明收到了信号
	<-ctx.Done()
	log.Info("Shutting down gracefully...")
	return nil
}
