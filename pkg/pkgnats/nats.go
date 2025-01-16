package pkgnats

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

func CreateNatsConnection(clientName, credentialsFileContent, natsUrl string) (*nats.Conn, error) {
	opts := []nats.Option{nats.Name(clientName)}
	opts = setupNatsConnectionOpts(opts)
	if len(credentialsFileContent) > 0 {
		opts = append(opts, nats.UserCredentials(credentialsFileContent))
		slog.Debug("pkgnats: Credentials file provided: " + credentialsFileContent)
	} else {
		slog.Debug("pkgnats: No credentials file provided")
	}
	slog.Debug("pkgnats: Creating nats connection with url: " + natsUrl + " and client name: " + clientName)
	nc, err := nats.Connect(natsUrl, opts...)
	if err != nil {
		slog.Debug("pkgnats: Error creating nats connection: " + err.Error())
		return nil, err
	}
	return nc, nil
}

func setupNatsConnectionOpts(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	slog.Debug("pkgnats: reconnect wait option set to " + reconnectDelay.String())
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	slog.Debug("pkgnats: max reconnects option set to " + fmt.Sprintf("%v", int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		slog.Debug(fmt.Sprintf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes()))
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		slog.Debug(fmt.Sprintf("Reconnected [%s]", nc.ConnectedUrl()))
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		slog.Debug(fmt.Sprintf("Exiting: %v", nc.LastError()))
	}))
	return opts
}
