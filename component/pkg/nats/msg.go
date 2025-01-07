package nats

import (
	jetstreamconsumer "github.com/Mattilsynet/map-providers/map-jetstream-nats/component/gen/mattilsynet/provider-jetstream-nats/jetstream-consumer"
	jetstreampublish "github.com/Mattilsynet/map-providers/map-jetstream-nats/component/gen/mattilsynet/provider-jetstream-nats/jetstream-publish"
	"github.com/Mattilsynet/map-providers/map-jetstream-nats/component/gen/mattilsynet/provider-jetstream-nats/types"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

type (
	Conn struct {
		js JetStreamContext
	}
	JetStreamContext struct{}
	Msg              struct {
		Subject string
		Reply   string
		Data    []byte
		Header  map[string][]string
	}
)

type MsgHandler func(msg *Msg)

func NewConn() *Conn {
	return &Conn{}
}

func (c *Conn) Jetstream() (*JetStreamContext, error) {
	return &c.js, nil
}

func (js *JetStreamContext) Publish(subj string, data []byte) cm.Result[string, struct{}, string] {
	return js.PublishMsg(&Msg{Subject: subj, Data: data})
}

func (js *JetStreamContext) PublishMsg(msg *Msg) cm.Result[string, struct{}, string] {
	jpMsg := jetstreampublish.Msg{
		Headers: toWitNatsHeaders(msg.Header),
		Data:    cm.ToList(msg.Data),
		Subject: msg.Subject,
	}
	return jetstreampublish.Publish(jpMsg)
}

func (js *JetStreamContext) Subscribe(msgHandler MsgHandler) {
	jetstreamconsumer.Exports.HandleMessage = toWitExportSubscription(msgHandler)
}

func toWitExportSubscription(msgHandler MsgHandler) func(msg jetstreamconsumer.Msg) cm.Result[string, struct{}, string] {
	return func(msg jetstreamconsumer.Msg) cm.Result[string, struct{}, string] {
		msgHandler(&Msg{
			Subject: msg.Subject,
			Reply:   msg.Reply,
			Data:    msg.Data.Slice(),
			Header:  toNatsHeaders(msg.Headers),
		})
		return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
	}
}

func toNatsHeaders(header cm.List[types.KeyValue]) map[string][]string {
	natsHeaders := make(map[string][]string)
	for _, kv := range header.Slice() {
		natsHeaders[kv.Key] = kv.Value.Slice()
	}
	return natsHeaders
}

func toWitNatsHeaders(header map[string][]string) cm.List[types.KeyValue] {
	keyValueList := make([]types.KeyValue, 0)
	for k, v := range header {
		keyValueList = append(keyValueList, types.KeyValue{
			Key:   k,
			Value: cm.ToList(v),
		})
	}
	return cm.ToList(keyValueList)
}
