package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	ps "github.com/mitchellh/go-ps"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type LastRunSummary struct {
	Time struct {
		Total    float64
		Last_run int
	}
	Events struct {
		Failure int
		Success int
		Total   int
	}
	Resources struct {
		Failed  int
		Skipped int
		Total   int
		Changed int
	}
}

type LastRunSummaryExporter struct {
	summary_file string
	mutex        sync.RWMutex

	time_total        prometheus.Gauge
	time_last_run     prometheus.Gauge
	events_failure    prometheus.Gauge
	events_success    prometheus.Gauge
	events_total      prometheus.Gauge
	resources_failed  prometheus.Gauge
	resources_skipped prometheus.Gauge
	resources_total   prometheus.Gauge
	resources_changed prometheus.Gauge
	agent_up          prometheus.Gauge
}

func SearchPuppetProc() float64 {
	processes, err := ps.Processes()
	if err != nil {
		log.Printf("error: %v", err)
		return 0
	}
	for _, process := range processes {
		if process.Executable() == "puppet" {
			return 1
		}
	}

	return 0
}

func (self *LastRunSummaryExporter) Collect(ch chan<- prometheus.Metric) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	summary_file, err := ioutil.ReadFile(self.summary_file)
	if err != nil {
		log.Printf("Couldn't read summary_file: %v", err)
	}

	last_run_summary := LastRunSummary{}
	err = yaml.Unmarshal([]byte(summary_file), &last_run_summary)
	if err != nil {
		log.Printf("error: %v", err)
	}

	self.time_total.Set(last_run_summary.Time.Total)
	self.time_last_run.Set(float64(last_run_summary.Time.Last_run))
	self.events_failure.Set(float64(last_run_summary.Events.Failure))
	self.events_success.Set(float64(last_run_summary.Events.Success))
	self.events_total.Set(float64(last_run_summary.Events.Total))
	self.resources_failed.Set(float64(last_run_summary.Resources.Failed))
	self.resources_skipped.Set(float64(last_run_summary.Resources.Skipped))
	self.resources_total.Set(float64(last_run_summary.Resources.Total))
	self.resources_changed.Set(float64(last_run_summary.Resources.Changed))
	self.agent_up.Set(SearchPuppetProc())

	ch <- self.time_total
	ch <- self.time_last_run
	ch <- self.events_failure
	ch <- self.events_success
	ch <- self.events_total
	ch <- self.resources_failed
	ch <- self.resources_skipped
	ch <- self.resources_total
	ch <- self.resources_changed
	ch <- self.agent_up
}

func (self *LastRunSummaryExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- self.time_total.Desc()
	ch <- self.time_last_run.Desc()
	ch <- self.events_failure.Desc()
	ch <- self.events_success.Desc()
	ch <- self.events_total.Desc()
	ch <- self.resources_failed.Desc()
	ch <- self.resources_skipped.Desc()
	ch <- self.resources_total.Desc()
	ch <- self.resources_changed.Desc()
	ch <- self.agent_up.Desc()
}

func CreateLastRunSummaryExporter(summary_file string) (*LastRunSummaryExporter, error) {
	return &LastRunSummaryExporter{
		summary_file: summary_file,

		time_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "time_total",
			Help:      "Last run duration ",
		}),
		time_last_run: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "time_last_run",
			Help:      "Last run timestamp ",
		}),

		events_failure: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "events_failure",
			Help:      "Last run events_failure ",
		}),
		events_success: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "events_success",
			Help:      "Last run events_success ",
		}),
		events_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "events_total",
			Help:      "Last run events_total ",
		}),

		resources_failed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "resources_failed",
			Help:      "Last run resources_failed ",
		}),
		resources_skipped: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "resources_skipped",
			Help:      "Last run resources_skipped ",
		}),
		resources_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "resources_total",
			Help:      "Last run resources_total ",
		}),
		resources_changed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "resources_changed",
			Help:      "Last run resources_changed ",
		}),
		agent_up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "puppet",
			Name:      "agent_up",
			Help:      "agent proccess up ",
		}),
	}, nil
}

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
	summary_file_def := "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"
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

	last_run_summary_exporter, _ := CreateLastRunSummaryExporter(summary_file)
	prometheus.MustRegister(last_run_summary_exporter)

	http.HandleFunc("/", RootHandler)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatalf("Can't bind on %v", listen)
	}
}
