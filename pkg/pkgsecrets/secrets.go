package secrets

import (
	"encoding/base64"
	"log/slog"

	"go.wasmcloud.dev/provider"
)

type Secrets struct {
	NatsCredentials string
}

func From(secretsMap map[string]provider.SecretValue) *Secrets {
	natsCredentialsEncrypted := secretsMap["nats-credentials"]
	natsCredentialsBase64Encoded := natsCredentialsEncrypted.Bytes.Reveal()
	natsCredentials := base64.StdEncoding.EncodeToString(natsCredentialsBase64Encoded)
	slog.Info("Nats credentials loaded, with size: ", "size", len(natsCredentials))
	return &Secrets{
		NatsCredentials: natsCredentials,
	}
}
