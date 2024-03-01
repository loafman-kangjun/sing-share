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
	return opts, err
}

func create(ctx context.Context) (*box.Box, error) {
	opts, err := readConfig()
	if err != nil {
		return nil, err
	}
	return box.New(box.Options{
		Context: ctx,
		Options: opts,
	})
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance, err := create(ctx)
	if err != nil {
		return err
	}

	// Setup signal handling
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(osSignals)

	go func() {
		<-osSignals
		cancel() // Signal the context to cancel
		if err := instance.Close(); err != nil {
			log.Error("Failed to close sing-box:", err)
		}
	}()

	if err := instance.Start(); err != nil {
		return err
	}

	// Wait for the context to be cancelled, indicating signal received
	<-ctx.Done()
	log.Info("Shutting down gracefully...")
	return nil
}
