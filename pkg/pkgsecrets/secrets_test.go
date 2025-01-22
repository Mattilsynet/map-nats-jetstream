package secrets

import (
	"log"
	"testing"

	"go.wasmcloud.dev/provider"
)

func TestFrom(t *testing.T) {
	log.Println("Hey")
	testJson := []byte("{\"kind\": \"String\", \"value\": \"dGVzdA==\"}")
	secretValueTest := provider.SecretValue{}
	secretValueTest.UnmarshalJSON(testJson)
	type args struct {
		secretsMap map[string]provider.SecretValue
	}
	tests := []struct {
		name string
		args args
		want *Secrets
	}{
		{
			"tc: adding no secrets, yields empty creds",
			args{},
			&Secrets{
				NatsCredentials: "",
			},
		},
		{
			"tc: adding secret base64encoded, yields creds",
			args{
				map[string]provider.SecretValue{
					"nats-credentials": secretValueTest,
				},
			},
			&Secrets{
				NatsCredentials: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.args.secretsMap); got.NatsCredentials != tt.want.NatsCredentials {
				t.Fatalf("From() = %v, want %v", got, tt.want)
			}
		})
	}
}
