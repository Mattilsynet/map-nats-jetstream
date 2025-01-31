# map-jetstream-nats
Supports very limited jetstream capabilities such as jetstream.Publish and registering a durable consumer to aggregate messages to a component.

## Link definition settings 
To configure this provider, use the following link settings in your link defitinion:

OBS! nats-credentials used in secrets in yaml examples underneath requires base64 encoded nats-credentials. The nats-credentials must be jwt and seed pem blocked.

```yaml
# setup your component to get a consumer, you have to implement 'jetstream-consumer' in your component
---
    - name: nats-jetstream
      type: capability
      properties:
        image: ghcr.io/Mattilsynet/map-nats-jetstream:v0.0.1-pre-17
      traits:
        - type: link
          properties:
            target: 
              name: <your-awesome-component-name>
            source:
              config:
                - name: nats-jetstream-nats-url
                  properties:
                    url: "nats://connect.nats.mattilsynet.io"
                - name: nats-jetstream-consumer-config
                  properties:
                    stream-name: "<some-stream-name>"
                    stream-retention-policy: "workqueue" # oneof "interest, workqueue, limits"
                    subject: "<your-awesome-subject"
                    durable-consumer-name: "<durable-consumer-name>"
              secrets:
                - name: nats-credentials
                  properties:
                    policy: nats-kv
                    key: <some-key-in-nats-kv>
            namespace: mattilsynet
            package: provider-jetstream-nats
            interfaces: [jetstream-consumer]
---
# setup your component to have jetstream.Publish capability
    - name: <your-awesome-component>
      type: component
      properties:
        image: <your-awesome-image>
        traits:
        - type: spreadscaler
          properties:
            instances: 1
        - type: link
          properties:
            target:
              name: map-jetstream-nats
              secrets:
                - name: nats-credentials
                  properties:
                    policy: nats-kv
                    key: <your-credentials-file-in-nats-kv>
              config:
                - name: nats-jetstream-config
                  properties:
                    url: "<your-nate-server>"
            namespace: mattilsynet
            package: provider-jetstream-nats
            interfaces: [jetstream-publish]
```
## Building

Prerequisites:

1. [Go toolchain](https://go.dev/doc/install)
1. [wit-bindgen-wrpc 0.7.0](https://github.com/bytecodealliance/wrpc), download the release binary
1. [wash](https://wasmcloud.com/docs/installation)

```bash
go generate ./...
go build .
```

Alternatively, you can generate, build and package this provider in one step:

```bash
wash build
```

You can build the included test component with `wash build -p ./component`.

## Running to test

Prerequisites:

1. [Go toolchain](https://go.dev/doc/install)
1. [nats-server](https://github.com/nats-io/nats-server)
1. [nats-cli](https://github.com/nats-io/natscli)

## Running as an application

You can deploy this provider, along with a [component](../component/) for testing, by deploying the [wadm.yaml](./wadm.yaml) application. Make sure to build the component with `wash build`.

```bash
# Build the component
cd component
wash build

# Return to the provider directory
cd ..

# Launch wasmCloud in the background
wash up -d
# Deploy the application
wash app deploy ./wadm.yaml
```
