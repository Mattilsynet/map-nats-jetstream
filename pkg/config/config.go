package config

type Config struct {
	NatsURL             string
	NatsSeed            string
	NatsJWT             string
	NatsUserCredentials string
}

func From(config map[string]string) Config {
	return Config{}
}
