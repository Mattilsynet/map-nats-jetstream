package main

import (

	// Go provider SDK

	"context"
	"errors"

	"github.com/Mattilsynet/map-jetstream-nats/bindings/mattilsynet/provider_jetstream_nats/types"
	"github.com/nats-io/nats.go"
	sdk "go.wasmcloud.dev/provider"
	wrpc "wrpc.io/go"
	wrpcnats "wrpc.io/go/nats"
)

// / Your PublishHandler struct is where you can store any state or configuration that your provider needs to keep track of.
type PublishHandler struct {
	// The provider instance
	provider *sdk.WasmcloudProvider
	// All components linked to this provider and their config.
	linkedFrom map[string]map[string]string
	// All components this provider is linked to and their config
	linkedTo        map[string]map[string]string
	natsConnections map[string]*nats.Conn
	js              map[string]nats.JetStreamContext
}

func (p *PublishHandler) Publish(ctx__ context.Context, msg *types.Msg) (*wrpc.Result[struct{}, string], error) {
	header, ok := wrpcnats.HeaderFromContext(ctx__)
	if !ok {
		p.provider.Logger.Warn("Received request from unknown origin")
		customErr := errors.New("received request from unknown origin")
		return wrpc.Err[struct{}](customErr.Error()), nil
	}
	sourceId := header.Get("source-id")
	js := p.js[sourceId]
	_, err := js.Publish(msg.Subject, msg.Data)
	if err != nil {
		return wrpc.Err[struct{}](err.Error()), nil
	}
	return wrpc.Ok[string](struct{}{}), nil
}

func NewPublishHandler(linkedFrom, linkedTo map[string]map[string]string) PublishHandler {
	return PublishHandler{
		linkedFrom:      linkedFrom,
		linkedTo:        linkedTo,
		natsConnections: make(map[string]*nats.Conn),
		js:              make(map[string]nats.JetStreamContext),
	}
}

func (p *PublishHandler) RegisterPublisherComponent(ctx context.Context, sourceId string) error {
	// read from secrets
	jwt := p.linkedFrom[sourceId]["jwt"]
	seed := p.linkedFrom[sourceId]["seed"]
	url := p.linkedFrom[sourceId]["url"]
	nc, natsConnErr := nats.Connect(url, nats.UserJWTAndSeed(jwt, seed))
	p.natsConnections[sourceId] = nc
	if natsConnErr != nil {
		return natsConnErr
	}
	js, jetStreamErr := nc.JetStream()
	if jetStreamErr != nil {
		return jetStreamErr
	}
	p.js[sourceId] = js
	return nil
}

func (p *PublishHandler) Shutdown() error {
	for _, nc := range p.natsConnections {
		nc.Close()
	}
	clear(p.natsConnections)
	clear(p.js)
	clear(p.linkedFrom)
	clear(p.linkedTo)
	return nil
}
