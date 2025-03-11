package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"prometheus-puppet-exporter/internal/exporter"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func main() {
	var listen string
	listen_def := "0.0.0.0:9140"
	pflag.StringVar(
		&listen,
		"listen",
		listen_def,
		"Listen address. Env LISTEN also can be used.",
	)

	var summary_file string
	summary_file_def := "/var/cache/puppet/public/last_run_summary.yaml"
	pflag.StringVar(
		&summary_file,
		"summary-file",
		summary_file_def,
		"Puppet last_run_summary.yaml. Env SUMMARY_FILE also can be used.",
	)
	pflag.Parse()

	if listen == listen_def && len(os.Getenv("LISTEN")) > 0 {
		listen = os.Getenv("LISTEN")
	}
	if summary_file == summary_file_def && len(os.Getenv("SUMMARY_FILE")) > 0 {
		summary_file = os.Getenv("SUMMARY_FILE")
	}

	log.Printf("Statring puppet agent exporter.")

	last_run_summary_exporter, _ := exporter.CreateLastRunSummaryExporter(summary_file)
	prometheus.MustRegister(last_run_summary_exporter)

	http.HandleFunc("/", RootHandler)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatalf("Can't bind on %v", listen)
	}
}
