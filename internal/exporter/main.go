package exporter

import (
	"io/ioutil"
	"log"
	"sync"

	fqdn "github.com/Showmax/go-fqdn"
	ps "github.com/mitchellh/go-ps"
	"github.com/prometheus/client_golang/prometheus"
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

	exporter_up       prometheus.Gauge
	time_total        prometheus.Gauge
	time_last_run     prometheus.Gauge
	events_failure    prometheus.Gauge
	events_success    prometheus.Gauge
	events_total      prometheus.Gauge
	resources_failed  prometheus.Gauge
	resources_skipped prometheus.Gauge
	resources_total   prometheus.Gauge
	resources_changed prometheus.Gauge
	summary_read_err  prometheus.Gauge
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

	self.exporter_up.Set(1)
	self.agent_up.Set(SearchPuppetProc())

	ch <- self.exporter_up
	ch <- self.agent_up

	summary_file, err := ioutil.ReadFile(self.summary_file)
	if err != nil {
		self.summary_read_err.Set(1)
		ch <- self.summary_read_err
		log.Printf("Couldn't read summary file: %v", err)
		return
	}

	last_run_summary := LastRunSummary{}
	err = yaml.Unmarshal([]byte(summary_file), &last_run_summary)
	if err != nil {
		self.summary_read_err.Set(1)
		ch <- self.summary_read_err
		log.Printf("Summary file unmarshal error: %v", err)
		return
	}
	if last_run_summary.Time.Last_run == 0 {
		self.summary_read_err.Set(1)
		ch <- self.summary_read_err
		log.Printf("Summary file unmarshal error: %v", err)
		return
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
	self.summary_read_err.Set(0)

	ch <- self.time_total
	ch <- self.time_last_run
	ch <- self.events_failure
	ch <- self.events_success
	ch <- self.events_total
	ch <- self.resources_failed
	ch <- self.resources_skipped
	ch <- self.resources_total
	ch <- self.resources_changed
	ch <- self.summary_read_err
}

func (self *LastRunSummaryExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- self.exporter_up.Desc()
	ch <- self.time_total.Desc()
	ch <- self.time_last_run.Desc()
	ch <- self.events_failure.Desc()
	ch <- self.events_success.Desc()
	ch <- self.events_total.Desc()
	ch <- self.resources_failed.Desc()
	ch <- self.resources_skipped.Desc()
	ch <- self.resources_total.Desc()
	ch <- self.resources_changed.Desc()
	ch <- self.summary_read_err.Desc()
	ch <- self.agent_up.Desc()
}

func CreateLastRunSummaryExporter(summary_file string) (*LastRunSummaryExporter, error) {
	hostname, err := fqdn.FqdnHostname()
	if err != nil {
		panic(err)
	}

	return &LastRunSummaryExporter{
		summary_file: summary_file,

		exporter_up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "exporter_up",
			Help:        "Exporter up ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		time_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "time_total",
			Help:        "Last run duration ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		time_last_run: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "time_last_run",
			Help:        "Last run timestamp ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),

		events_failure: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "events_failure",
			Help:        "Last run events_failure ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		events_success: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "events_success",
			Help:        "Last run events_success ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		events_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "events_total",
			Help:        "Last run events_total ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),

		resources_failed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "resources_failed",
			Help:        "Last run resources_failed ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		resources_skipped: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "resources_skipped",
			Help:        "Last run resources_skipped ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		resources_total: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "resources_total",
			Help:        "Last run resources_total ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		resources_changed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "resources_changed",
			Help:        "Last run resources_changed ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		summary_read_err: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "summary_read_err",
			Help:        "Summary file read error ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
		agent_up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "puppet",
			Name:        "agent_up",
			Help:        "agent proccess up ",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}),
	}, nil
}
