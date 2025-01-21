package secrets

import (
	"encoding/base64"

	"go.wasmcloud.dev/provider"
)

type Secrets struct {
	NatsCredentials string
}

func From(secretsMap map[string]provider.SecretValue) *Secrets {
	natsCredentialsEncrypted := secretsMap["nats-credentials"]
	natsCredentialsBase64Encoded := natsCredentialsEncrypted.Bytes.Reveal()
	natsCredentials := base64.StdEncoding.EncodeToString(natsCredentialsBase64Encoded)
	return &Secrets{
		NatsCredentials: natsCredentials,
	}
}
