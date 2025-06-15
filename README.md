# idle-svc

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**idle-svc** is a tiny command-line tool that scans your Kubernetes cluster and lists every Service whose EndpointSlice or legacy Endpoints object contains **zero _ready_ addresses**.

These "idle" (or "orphan") Services still hold a ClusterIP and DNS record even though no Pods back them.

---

## Features

- **Read-only**: requires only `get` and `list` permissions on the `services`, `endpointslices` and `endpoints` resources
- Handles both EndpointSlice (Kubernetes ≥ 1.21) **and** legacy Endpoints
- Human-friendly table output, or JSON with `--json`
- `--exit-code` flag lets CI pipelines fail when idle Services exist

---

## Deployment options

`idle-svc` is packaged four different ways—use whichever fits your workflow.

1. **CLI binary**  
   Build with `make build` (or `go install github.com/0xabrar/idle-svc@latest`) and run interactively or in CI:
   ```bash
   idle-svc -A --exit-code     # fail pipeline when orphans exist
   ```

2. **Docker container**  
   Pull `ghcr.io/0xabrar/idle-svc:<tag>` and either:
   * Run locally, mounting your `$HOME/.kube/config`.
   * Add it as a **sidecar** in any Deployment to expose `/metrics` for Prometheus:
     ```yaml
     containers:
       - name: app
         image: my-team/app:1.0
       - name: idle-svc
         image: ghcr.io/0xabrar/idle-svc:latest
         args: ["-A", "--listen", ":9090", "--watch"]
     ```
     The sidecar continuously scans the cluster and updates the gauge `idle_services_total` without scheduling a separate CronJob.

3. **Helm chart**  
   One-command install into any cluster.  Creates:
   * A ServiceAccount + read-only Role/RoleBinding
   * A CronJob that runs every 15 min (configurable) and emits metrics
   * An optional Service (port 9090) for scraping

4. **Go library**  
   Import just the detection logic in another controller/operator:
   ```go
   import "github.com/0xabrar/idle-svc/pkg/orphanfinder"

   orphans, _ := orphanfinder.Find(ctx, coreClient, discoClient, "", true)
   if len(orphans) > 0 { /* … */ }
   ```

Pick the lightest option that solves your problem: ad-hoc scans (CLI), continuous monitoring (Helm or sidecar), or embedding (library).

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

## Container image

A multi-arch (amd64/arm64) image is published on every release:

```bash
docker run --rm ghcr.io/0xabrar/idle-svc:latest -A --listen :9090 --watch
```

The container runs as **non-root** and needs only the Kubernetes config in the pod. When deployed in-cluster use the Helm chart below.

---

## Helm chart

Install the scheduled scan (CronJob) into the `monitoring` namespace:

```bash
helm repo add idle-svc https://0xabrar.github.io/idle-svc-charts
helm install idle-svc idle-svc/idle-svc \
  --namespace monitoring --create-namespace \
  --set image.tag=$(git describe --tags)
```

By default the chart:
- Runs every 15 minutes
- Scans all namespaces (`-A`)
- Exposes Prometheus metrics on port 9090

Flags can be overridden via `values.yaml` (`extraArgs`).

---

## Usage (CLI)

| Command | Purpose |
|---------|---------|
| `idle-svc` | One-shot scan of the current namespace |
| `idle-svc -A` | One-shot scan of all namespaces |
| `idle-svc --namespace demo` | Scan only the `demo` namespace |
| `idle-svc -A --json` | Output JSON instead of table |
| `idle-svc -A --exit-code` | Exit 1 if idle Services exist (CI) |
| `idle-svc -A --listen :9090 --watch` | Continuous scan every 30 s and expose Prometheus metrics |
| `idle-svc --interval 5m --watch` | Scan every 5 minutes |

---

### Prometheus metric

When `--listen` is supplied, the binary starts an HTTP server and updates a gauge:

```
# HELP idle_services_total Number of Services with zero ready endpoints
# TYPE idle_services_total gauge
idle_services_total{namespace="demo"} 1
```

Create an alert:

```
- alert: OrphanedServicesExist
  expr: idle_services_total > 0
  for: 15m
  labels:
    severity: warning
  annotations:
    summary: "Idle Services detected"
```

For full YAML examples see [`docs/rbac.yaml`](docs/rbac.yaml) and [`docs/alerting.md`](docs/alerting.md).

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

---

## License

Licensed under the [Apache License 2.0](LICENSE).

---

## When to use idle-svc vs. Prometheus/kube-state-metrics

Some clusters already run Prometheus _and_ kube-state-metrics (KSM). With the right PromQL you **can** detect idle Services there, so why bother with another tool?

| Scenario | KSM + custom PromQL | idle-svc |
|----------|--------------------|----------|
| Local dev kind/minikube cluster | Need to install Prometheus + KSM first | One `go run`/`docker run` away |
| CI gate (fail PR on idle Service) | Evaluate PromQL via `curl` or custom script | `idle-svc -A --exit-code` does it |
| Air-gapped prod cluster w/o Prometheus | Not possible | CronJob with 15 MB static binary |
| Want a ready-made metric | Write/maintain alert rule | `--listen :9090` exposes `idle_services_total` |
| Need a Go library to embed in controller | N/A | `pkg/orphanfinder` |

idle-svc is **not** the best choice if:

* You already have Prometheus + KSM, and
* You are happy to maintain a PromQL rule such as:
  ```promql
  kube_service_info * 0
  # + custom logic to ensure kube_endpoint_address_available == 0
  ```
* You don't need CLI/CI integration or auto-deletion features.

For everyone else—especially small clusters, developer laptops, or CI pipelines—`idle-svc` provides a zero-setup, purpose-built solution that can still integrate with Prometheus when available.

---

## Maintainers

For release instructions (building images, publishing the Helm chart, tagging versions), see [`docs/publishing.md`](docs/publishing.md).

---
