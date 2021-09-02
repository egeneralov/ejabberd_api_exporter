package main

import (
	"flag"
	"github.com/egeneralov/ejabberd_api_exporter/internal/api"
	"github.com/egeneralov/ejabberd_api_exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	vhost     = flag.String("vhost", "ejabberd", "virtual host to use")
	endpoint  = flag.String("endpoint", "https://ejabberd:5443", "endpoint to go")
	bind      = flag.String("bind", "0.0.0.0:8080", "bind to")
	namespace = flag.String("namespace", "ejabberd", "metric namespace")
)

func main() {
	flag.Parse()
	prometheus.MustRegister(collector.New(
		api.New(
			*vhost,
			*endpoint,
		),
		*namespace,
	))
	err := http.ListenAndServe(*bind, promhttp.Handler())
	if err != nil {
		panic(err)
	}
}
