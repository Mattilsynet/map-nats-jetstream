package main

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/Mattilsynet/map-jetstream-nats/bindings/mattilsynet/provider_jetstream_nats/jetstream_consumer"
	"github.com/Mattilsynet/map-jetstream-nats/bindings/mattilsynet/provider_jetstream_nats/types"
	"github.com/Mattilsynet/map-jetstream-nats/pkg/config"
	"github.com/Mattilsynet/map-jetstream-nats/pkg/pkgnats"
	secrets "github.com/Mattilsynet/map-jetstream-nats/pkg/pkgsecrets"
	"github.com/nats-io/nats.go"
	sdk "go.wasmcloud.dev/provider"
)

// / Your Handler struct is where you can store any state or configuration that your provider needs to keep track of.
type ConsumeHandler struct {
	// The provider instance
	provider *sdk.WasmcloudProvider
	// All components linked to this provider and their config.
	linkedFrom map[string]map[string]string
	// All components this provider is linked to and their config
	linkedTo        map[string]map[string]string
	natsConnections map[string]*nats.Conn
	subscriptions   map[string][]*nats.Subscription
}

func NewConsumeHandler(linkedFrom, linkedTo map[string]map[string]string) ConsumeHandler {
	return ConsumeHandler{
		linkedFrom:      linkedFrom,
		linkedTo:        linkedTo,
		natsConnections: make(map[string]*nats.Conn),
		subscriptions:   make(map[string][]*nats.Subscription),
	}
}

func (p *ConsumeHandler) RegisterConsumerComponent(target string, config *config.Config, secrets *secrets.Secrets) error {
	streamRetentionPolicy := config.StreamRetentionPolicy
	slog.Info("1")
	var retentionPolicy nats.RetentionPolicy
	switch streamRetentionPolicy {
	case "limits":
		retentionPolicy = nats.LimitsPolicy
	case "interest":
		retentionPolicy = nats.InterestPolicy
	case "workqueue":
		retentionPolicy = nats.WorkQueuePolicy
	default:
		return errors.New("invalid retention policy" + streamRetentionPolicy)

	}
	slog.Info("11")
	// INFO: Might not need the line of code underneath, doing it for backwards compatability with previous commit
	p.linkedFrom[target] = config.ProviderConfig
	// TODO: Put all the nats configuration inside pkgnats and leave this clean and only about how the component will get msgs
	slog.Info("111")
	streamName := config.StreamName
	// read from secrets
	durableConsumerName := config.ConsumerName
	credentialsFile := secrets.NatsCredentials
	subject := config.Subject
	url := config.NatsURL
	slog.Info("1111")
	nc, natsConnErr := pkgnats.CreateNatsConnection(target, credentialsFile, url)
	p.natsConnections[target] = nc
	if natsConnErr != nil {
		return natsConnErr
	}
	slog.Info("11111")
	js, jsErr := nc.JetStream()
	if jsErr != nil {
		return jsErr
	}
	streamInfo, _ := js.StreamInfo(streamName)
	if streamInfo == nil {
		_, b := js.AddStream(&nats.StreamConfig{Name: streamName, Subjects: []string{subject}, Retention: retentionPolicy})
		if b != nil {
			return b
		}
	} else {
		if !slices.Contains(streamInfo.Config.Subjects, subject) {
			streamInfo.Config.Subjects = append(streamInfo.Config.Subjects, subject)
			_, b := js.UpdateStream(&streamInfo.Config)
			if b != nil {
				return b
			}
		}
	}
	client := p.provider.OutgoingRpcClient(target)
	p.subscriptions[target] = make([]*nats.Subscription, 0)

	sub, subscriptionErr := js.Subscribe(subject, func(m *nats.Msg) {
		headers := convertMapToWitHeaders(m.Header)
		msg := &types.Msg{
			Data:    m.Data,
			Reply:   m.Reply,
			Subject: m.Subject,
			Headers: headers,
		}
		ctx := context.Background()
		str, err := jetstream_consumer.HandleMessage(ctx, client, msg)
		if err != nil || str.Err != nil {
			p.provider.Logger.Error("error handling message", "err", err)
			m.Nak(nats.AckWait(1 * time.Second))
		}
		m.Ack()
	}, nats.Durable(durableConsumerName), nats.BindStream(streamName))
	if subscriptionErr != nil {
		p.provider.Logger.Error("Erroring: ", "error", subscriptionErr)
		nc.Close()
		return subscriptionErr
	}
	p.subscriptions[target] = append(p.subscriptions[target], sub)
	return nil
}

func (p *ConsumeHandler) DelSourceLink(target string, sourceId string) {
	for _, sub := range p.subscriptions[target] {
		sub.Unsubscribe()
	}
	delete(p.subscriptions, target)
	p.natsConnections[target].Close()
	delete(p.linkedTo, sourceId)
	delete(p.natsConnections, target)
	delete(p.linkedFrom, target)
}

func convertMapToWitHeaders(header nats.Header) []*types.KeyValue {
	headers := make([]*types.KeyValue, 0)
	for k, v := range header {
		headers = append(headers, &types.KeyValue{
			Key:   k,
			Value: v,
		})
	}
	return headers
}

func (p *ConsumeHandler) Shutdown() error {
	for target := range p.subscriptions {
		p.DelSourceLink(target, "")
	}
	clear(p.natsConnections)
	clear(p.linkedFrom)
	clear(p.linkedTo)
	clear(p.subscriptions)
	return nil
}
