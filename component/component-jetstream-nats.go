//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world component --out gen ./wit
package main

import (
	"github.com/Mattilsynet/map-providers/map-jetstream-nats/component/pkg/nats"
	"go.wasmcloud.dev/component/log/wasilog"
)

var (
	nc *nats.Conn
	js *nats.JetStreamContext
)

func init() {
	nc = nats.NewConn()
	var err error
	js, err = nc.Jetstream()
	if err != nil {
		return
	}
	js.Subscribe(msgHandler)
}

func main() {}

func msgHandler(msg *nats.Msg) {
	logger := wasilog.ContextLogger("component-jetstream-nats-handler")
	logger.Info("message received", "subject", msg.Subject, "data", string(msg.Data))
	js.Publish("test2.response", []byte("hey back"))
}
