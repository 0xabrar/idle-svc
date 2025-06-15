# Prometheus Alerting for idle-svc

When you run `idle-svc` with `--listen :9090` (or via the Helm chart), it exposes the following metric:

```
# HELP idle_services_total Number of Services with zero ready endpoints
# TYPE idle_services_total gauge
idle_services_total{namespace="default"} 1
```

A minimal `PrometheusRule` to fire after 15 minutes of continuous idleness:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: idle-svc-alerts
  labels:
    prometheus: kube-prometheus # adjust to match your stack
spec:
  groups:
    - name: idle-svc
      rules:
        - alert: OrphanedServicesExist
          expr: idle_services_total > 0
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Idle Services detected"
            description: |
              {{ $value }} idle Service(s) have zero ready endpoints.
              Run `idle-svc -A` or open the dashboard for details.
``` 