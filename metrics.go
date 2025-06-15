package main

import (
    "net/http"

    "github.com/0xabrar/idle-svc/pkg/orphanfinder"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var idleGauge = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "idle_services_total",
        Help: "Number of Services with zero ready endpoints",
    },
    []string{"namespace"},
)

func initMetrics(listenAddr string) {
    prometheus.MustRegister(idleGauge)
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(listenAddr, nil)
}

func updateMetrics(orphans []orphanfinder.Orphan) {
    // Reset then increment per namespace count
    idleGauge.Reset()
    nsCount := make(map[string]float64)
    for _, o := range orphans {
        nsCount[o.Namespace]++
    }
    for ns, cnt := range nsCount {
        idleGauge.WithLabelValues(ns).Set(cnt)
    }
} 