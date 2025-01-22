package secrets

import (
	"testing"
)

func TestFrom(t *testing.T) {
	test := "dGVzdA=="
	type args struct {
		string
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
				test,
			},
			&Secrets{
				NatsCredentials: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.args.string); got.NatsCredentials != tt.want.NatsCredentials {
				t.Fatalf("From() = %v, want %v", got, tt.want)
			}
		})
	}
}
