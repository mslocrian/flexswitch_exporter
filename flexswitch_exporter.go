// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mslocrian/flexswitch_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	defaultCollectors = "ports"
)

var (
	scrapeDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: collector.Namespace,
			Subsystem: "exporter",
			Name:      "scrape_duration_seconds",
			Help:      "flexswitch_exporter: Duration of a scrape job.",
		},
		[]string{"collector", "result"},
	)

	flexswitchRequestErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "flexswitch_request_errors_total",
			Help: "Errors in requests to the FlexSwitch exporter",
		},
	)

	enabledCollectors = flag.String("collectors.enabled", defaultCollectors, "Comma-separated list of collectors to use.")

	configFile = flag.String("config.file", "flexswitch.yml",
		"Path to configuration file",
	)
	showVersion   = flag.Bool("version", false, "Print version information.")
	listenAddress = flag.String("web.listen-address", ":9117", "Address on which to expose metrics and web interface.")
)

// NodeCollector implements the prometheus.Collector interface.
type NodeCollector struct {
	collectors      map[string]collector.Collector
	collectorParams collector.FlexSwitchParams
}

/*
type FlexSwitchParams struct {
	target string
	proto string
	port int
	username string
	password string
}
*/

// Describe implements the prometheus.Collector interface.
func (n NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	scrapeDurations.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (n NodeCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.collectors))
	for name, c := range n.collectors {
		go func(name string, c collector.Collector) {
			execute(name, n.collectorParams, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
	scrapeDurations.Collect(ch)
}

func execute(name string, params collector.FlexSwitchParams, c collector.Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(params, ch)
	duration := time.Since(begin)
	var result string

	if err != nil {
		log.Errorf("ERROR: %s collector failed after %fs: %s", name, duration.Seconds(), err)
		result = "error"
	} else {
		log.Debugf("OK: %s collector succeeded after %fs.", name, duration.Seconds())
		result = "success"
	}
	scrapeDurations.WithLabelValues(name, result).Observe(duration.Seconds())
}

func loadCollectors(list string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}
	for _, name := range strings.Split(list, ",") {
		fn, ok := collector.Factories[name]
		if !ok {
			return nil, fmt.Errorf("collector '%s' not available", name)
		}
		c, err := fn()
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}
	return collectors, nil
}

func init() {
	prometheus.MustRegister(scrapeDurations)
	prometheus.MustRegister(flexswitchRequestErrors)
	prometheus.MustRegister(version.NewCollector("flexswitch_exporter"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	collectors, err := loadCollectors(*enabledCollectors)
	if err != nil {
		log.Fatalf("Couldn't load collectors: %s", err)
	}

	cfg, err := LoadFile(*configFile)

	if err != nil {
		msg := fmt.Sprintf("Error parsing config file: %s", err)
		http.Error(w, msg, 400)
		log.Error(msg)
		return
	}

	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		flexswitchRequestErrors.Inc()
		return
	}

	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		moduleName = "default"
	}
	module, ok := (*cfg)[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module '%s'", moduleName), 400)
		flexswitchRequestErrors.Inc()
		return
	}
	log.Debug("Scraping target '%s' with module '%s'", target, moduleName)

	fsp := collector.FlexSwitchParams{Target: target,
		Proto:    module.Proto,
		Port:     module.Port,
		Username: module.Auth.Username,
		Password: module.Auth.Password}

	nodeCollector := NodeCollector{collectors: collectors,
		collectorParams: fsp}

	registry := prometheus.NewRegistry()
	registry.MustRegister(nodeCollector)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("flexswitch_exporter"))
		os.Exit(0)
	}

	_, err := LoadFile(*configFile)
	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	log.Infoln("Starting flexswitch_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/flexswitch", handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>FlexSwitch Exporter</title></head>
			<body>
			<h1>Node Exporter</h1>
			<p><a href="/flexswitch">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
