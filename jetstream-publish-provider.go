package main

import (

	// Go provider SDK

	"context"
	"errors"
	"fmt"

	"github.com/Mattilsynet/map-nats-jetstream/bindings/mattilsynet/map_nats_jetstream/types"
	"github.com/Mattilsynet/map-nats-jetstream/pkg/config"
	"github.com/Mattilsynet/map-nats-jetstream/pkg/pkgnats"
	secrets "github.com/Mattilsynet/map-nats-jetstream/pkg/pkgsecrets"
	"github.com/nats-io/nats.go"
	sdk "go.wasmcloud.dev/provider"
	wrpc "wrpc.io/go"
	wrpcnats "wrpc.io/go/nats"
)

const JETSTREAM_PROVIDER_NAME = "JETSTREAM_PUBLISHER"

// / Your PublishHandler struct is where you can store any state or configuration that your provider needs to keep track of.
type PublishHandler struct {
	// The provider instance
	provider *sdk.WasmcloudProvider
	// All components linked to this provider and their config.
	linkedFrom map[string]map[string]string
	// All components this provider is linked to and their config
	linkedTo        map[string]map[string]string
	natsConnections map[string]*nats.Conn
}

func (p *PublishHandler) Publish(ctx__ context.Context, msg *types.Msg) (*wrpc.Result[struct{}, string], error) {
	header, ok := wrpcnats.HeaderFromContext(ctx__)
	if !ok {
		p.provider.Logger.Warn("Received request from unknown origin")
		customErr := errors.New("received request from unknown origin")
		return wrpc.Err[struct{}](customErr.Error()), nil
	}
	p.provider.Logger.Info("got request from source-id: ", "source-id", header.Get("source-id"))
	sourceId := header.Get("source-id")
	natsConn, ok := p.natsConnections[sourceId]
	if !ok {
		return nil, fmt.Errorf("no nats connection found for source-id: %s", sourceId)
	}
	js, err := natsConn.JetStream()
	if err != nil {
		p.provider.Logger.Error("Error getting JetStream context: ", "err", err)
		return nil, err
	}
	_, err = js.Publish(msg.Subject, msg.Data)
	if err != nil {
		p.provider.Logger.Error("Error publishing message: ", "err", err)
		return wrpc.Err[struct{}](err.Error()), nil
	}
	return wrpc.Ok[string](struct{}{}), nil
}

func NewPublishHandler(linkedFrom, linkedTo map[string]map[string]string) PublishHandler {
	return PublishHandler{
		linkedFrom:      linkedFrom,
		linkedTo:        linkedTo,
		natsConnections: make(map[string]*nats.Conn),
	}
}

func (p *PublishHandler) RegisterPublisherComponent(ctx context.Context, sourceId string, config *config.Config, secrets *secrets.Secrets) error {
	url := config.NatsURL
	p.linkedFrom[sourceId] = config.ProviderConfig
	nc, natsConnErr := pkgnats.CreateNatsConnection(sourceId, secrets.NatsCredentials, url)
	if natsConnErr != nil {
		p.provider.Logger.Error("Error creating nats connection: ", "err", natsConnErr)
		return natsConnErr
	}
	p.natsConnections[sourceId] = nc
	return nil
}

func (p *PublishHandler) DelTargetLink(target string, sourceId string) {
	p.natsConnections[target].Close()
	delete(p.linkedFrom, target)
	delete(p.natsConnections, target)
}

func (p *PublishHandler) Shutdown() error {
	for _, nc := range p.natsConnections {
		nc.Close()
	}
	clear(p.natsConnections)
	clear(p.linkedFrom)
	clear(p.linkedTo)
	return nil
}
