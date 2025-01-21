//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/Mattilsynet/map-jetstream-nats/bindings wit

package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"

	server "github.com/Mattilsynet/map-jetstream-nats/bindings"
	"github.com/Mattilsynet/map-jetstream-nats/pkg/config"
	secrets "github.com/Mattilsynet/map-jetstream-nats/pkg/pkgsecrets"
	"go.wasmcloud.dev/provider"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Initialize the provider with callbacks to track linked components
	publishHandler := NewPublishHandler(
		make(map[string]map[string]string),
		make(map[string]map[string]string),
	)
	consumeHandler := NewConsumeHandler(
		make(map[string]map[string]string),
		make(map[string]map[string]string),
	)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	p, err := provider.New(
		provider.SourceLinkPut(func(link provider.InterfaceLinkDefinition) error {
			return handleNewConsumerComponent(&consumeHandler, link)
		}),
		provider.TargetLinkPut(func(link provider.InterfaceLinkDefinition) error {
			return handleNewTargetLink(&publishHandler, link)
		}),
		provider.SourceLinkDel(func(link provider.InterfaceLinkDefinition) error {
			return handleDelConsumerComponent(&consumeHandler, link)
		}),
		provider.TargetLinkDel(func(link provider.InterfaceLinkDefinition) error {
			return handleDelPublishConsumer(&publishHandler, link)
		}),
		provider.HealthCheck(func() string {
			return handleHealthCheck(&publishHandler, &consumeHandler)
		}),
		provider.Shutdown(func() error {
			return handleShutdown(&publishHandler, &consumeHandler)
		}),
	)
	if err != nil {
		return err
	}
	// Store the provider for use in the handlers
	publishHandler.provider = p
	consumeHandler.provider = p
	p.Logger = logger
	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(p.RPCClient, &publishHandler)
	if err != nil {
		p.Shutdown()
		return err
	}

	// Handle control interface operations
	go func() {
		err := p.Start()
		providerCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	select {
	case err = <-providerCh:
		stopFunc()
		return err
	case <-signalCh:
		p.Shutdown()
		stopFunc()
	}

	return nil
}

func handleNewConsumerComponent(consumeHandler *ConsumeHandler, link provider.InterfaceLinkDefinition) error {
	slog.Info("Handling new source link", "link", link)
	if slices.Contains(link.Interfaces, "jetstream-consumer") {
		consumeHandler.linkedFrom[link.Target] = link.SourceConfig
		consumerConfig := config.From(link.SourceConfig)
		slog.Info("consumer link", "link", link)
		slog.Info("consumerConfig", "consumerConfig", consumerConfig)
		secrets := secrets.From(link.SourceSecrets)
		// TODO: put it in consumerHandler
		err := consumeHandler.RegisterConsumerComponent(link.Target, consumerConfig, secrets)
		if err != nil {
			slog.Error("exiting with", "error", err)
			os.Exit(1)
		}
	}
	return nil
}

func handleNewTargetLink(publishHandler *PublishHandler, link provider.InterfaceLinkDefinition) error {
	slog.Info("Handling new target link", "link", link)
	if slices.Contains(link.Interfaces, "jetstream-publish") {
		publisherConfig := config.From(link.TargetConfig)
		publisherSecrets := secrets.From(link.SourceSecrets)
		publishHandler.RegisterPublisherComponent(context.Background(), link.SourceID, publisherConfig, publisherSecrets)
	}
	return nil
}

func handleDelConsumerComponent(consumeHandler *ConsumeHandler, link provider.InterfaceLinkDefinition) error {
	slog.Info("Handling del source link", "link", link)
	consumeHandler.DelSourceLink(link.Target, link.SourceID)
	return nil
}

func handleDelPublishConsumer(publishHandler *PublishHandler, link provider.InterfaceLinkDefinition) error {
	slog.Info("Handling del target link", "link", link)
	delete(publishHandler.linkedFrom, link.Target)
	publishHandler.DelTargetLink(link.Target, link.SourceID)
	return nil
}

func handleHealthCheck(publishHandler *PublishHandler, consumeHandler *ConsumeHandler) string {
	return "provider healthy"
}

func handleShutdown(handler *PublishHandler, consumeHandler *ConsumeHandler) error {
	handler.Shutdown()
	consumeHandler.Shutdown()
	return nil
}
