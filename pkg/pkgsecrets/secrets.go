package secrets

import (
	"encoding/base64"
	"log/slog"
)

type Secrets struct {
	NatsCredentials string
}

func From(base64encoded string) *Secrets {
	slog.Info("Nats credentials base64encoded, with size: ", "size", len(base64encoded))
	natsCredentials, err := base64.StdEncoding.DecodeString(base64encoded)
	if err != nil {
		return nil
	}
	slog.Info("Nats credentials loaded, with size: ", "size", len(natsCredentials))
	return &Secrets{
		NatsCredentials: string(natsCredentials),
	}
}
