package config

type Config struct {
	NatsURL               string
	StreamName            string
	Subject               string
	StreamRetentionPolicy string
	ConsumerName          string
	ProviderConfig        map[string]string
}

func From(config map[string]string) *Config {
	return &Config{
		NatsURL:               config["url"],
		StreamName:            config["stream-name"],
		StreamRetentionPolicy: config["stream-retention-policy"],
		ConsumerName:          config["durable-consumer-name"],
		Subject:               config["subject"],
		ProviderConfig:        config,
	}
}
