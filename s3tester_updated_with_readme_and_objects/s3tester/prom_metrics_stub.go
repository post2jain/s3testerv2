package main

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    OpsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "s3tester_ops_total",
            Help: "Total operations executed",
        },
        []string{"op"},
    )
    OpsErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "s3tester_ops_errors_total",
            Help: "Total operation errors",
        },
        []string{"op", "code"},
    )
)

func init() {
    prometheus.MustRegister(OpsTotal)
    prometheus.MustRegister(OpsErrors)
}

func StartMetricsServer(addr string) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    srv := &http.Server{Addr: addr, Handler: mux}
    go func() {
        _ = srv.ListenAndServe()
    }()
    return srv
}
