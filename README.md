# idle-svc

[![License](https://img.shields.io/github/license/0xabrar/idle-svc.svg)](LICENSE)

**idle-svc** is a tiny command-line tool that scans your Kubernetes cluster and lists every Service whose EndpointSlice or legacy Endpoints object contains **zero _ready_ addresses**.

These "idle" (or "orphan") Services still hold a ClusterIP and DNS record even though no Pods back them.

---

## Features

- **Read-only**: requires only `get`, `list` and `watch` permissions on the `services`, `endpointslices` and `endpoints` resources
- Handles both EndpointSlice (Kubernetes ≥ 1.21) **and** legacy Endpoints
- Human-friendly table output, or JSON with `--json`
- `--exit-code` flag lets CI pipelines fail when idle Services exist

---

## Installation

Prerequisites: **Go 1.22** or newer must be installed and on your `PATH`.

- Clone this repository:

```bash
git clone https://github.com/0xabrar/idle-svc.git
cd idle-svc
```

- Build the binary locally:

```bash
make build   # produces ./idle-svc
```

- Optionally install it into your Go bin directory:

```bash
make install # copies the binary to $(go env GOBIN)
```

Ensure that `$(go env GOBIN)` (typically `$HOME/go/bin`) is on your `PATH` so you can invoke `idle-svc` from anywhere.

---

## Usage

| Command | Purpose |
|---------|---------|
| `idle-svc` | Scan the current namespace |
| `idle-svc -A` | Scan **all** namespaces |
| `idle-svc --namespace demo` | Scan only the `demo` namespace |
| `idle-svc -A --json --exit-code` | Output JSON _and_ return **exit 1** if any idle Services are found |

Typical output when an idle Service exists:

```text
NAMESPACE   SERVICE   TYPE        CLUSTER-IP     AGE
example     ghost     ClusterIP   10.96.62.17    3h12m
```

If no idle Services exist:

```text
👍  no idle services found
```

---

## Quick local test with kind

```bash
# create a local Kubernetes cluster
kind create cluster --name idle-svc-demo

# create a Deployment and a Service
git clone https://github.com/kubernetes/examples.git
kubectl apply -f examples/service/deployment.yaml
kubectl apply -f examples/service/service.yaml

# scale the Deployment to zero replicas to orphan the Service
kubectl scale deployment nginx-deployment --replicas 0

# run idle-svc – the orphaned Service should appear in the output
idle-svc -A
```
