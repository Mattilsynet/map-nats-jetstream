# map-jetstream-nats
Supports very limited jetstream capabilities such as jetstream.Publish and registering a durable consumer to aggregate messages to a component.
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

## Customizing

Customizing this provider to meet your needs of a custom capability takes just a few steps.

1. Update the [wit/world.wit](./wit/world.wit) to include the data types and functions that model your custom capability. You can use the example as a base and the [component model WIT reference](https://component-model.bytecodealliance.org/design/wit.html) as a guide for types and keywords.
1. Implement any provider `export`s in [provider.go](./provider.go) as methods of your `Handler`.
1. Use any provider `import`s in [provider.go](./provider.go) to invoke linked components. Check out the `Call()` function for an example for how to invoke a component using RPC.
