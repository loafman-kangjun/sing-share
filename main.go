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
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func readConfig() (option.Options, error) {
    var opts option.Options
    configContent, err := os.ReadFile("config.json")
    if err != nil {
        return opts, err
    }
    err = json.Unmarshal(configContent, &opts)
    if err != nil {
        return opts, err
    }
    return opts, nil
}

func run() error {
    // 创建一个可以被取消的上下文
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 读取配置文件
    opts, err := readConfig()
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
