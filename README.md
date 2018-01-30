controller-gen
==============

Generate Kubernetes controller stubs that sync configurable resource types

## Installation

1. Install Go 1.9+
2. `go get github.com/llparse/controller-gen`

## Example Usage

Generate the controller code

```sh
controller-gen \
  --output-package "github.com/llparse/controller-gen/example_generated" \
  --go-header-file /dev/null \
  -v 3 \
  --name example \
  --types core/v1/Pod,core/v1/Service,apps/v1beta2/Deployment,storage/v1/StorageClass
```

Build the controller

```sh
go build -o controller-example example_generated/cmd/controller/example/main.go
```

Run it!

```sh
./controller-example -kubeconfig ~/.kube/config -v 5
```

