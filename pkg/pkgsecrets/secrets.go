package secrets

import "go.wasmcloud.dev/provider"

type Secrets struct {
	NatsCredentials string
}

func From(secretsMap map[string]provider.SecretValue) *Secrets {
	natsCredentialsEncrypted := secretsMap["nats-credentials"]
	natsCredentials := natsCredentialsEncrypted.String.Reveal()
	return &Secrets{
		NatsCredentials: natsCredentials,
	}
}
