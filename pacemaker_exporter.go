// Copyright 2018 Mario Trangoni
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
	"fmt"
	"net/http"

	"github.com/mjtrangoni/pacemaker_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector("pacemaker_exporter"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var num int

	filters := r.URL.Query()["collect[]"]
	log.Debugln("collect query:", filters)

	nc, err := collector.NewPacemakerCollector()
	if err != nil {
		log.Warnln("Couldn't create", err)
		w.WriteHeader(http.StatusBadRequest)

		num, err = w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))

		if err != nil {
			log.Fatal(num, err)
		}

		return
	}

	registry := prometheus.NewRegistry()
	err = registry.Register(nc)

	if err != nil {
		log.Errorln("Couldn't register collector:", err)
		w.WriteHeader(http.StatusInternalServerError)

		num, err = w.Write([]byte(fmt.Sprintf("Couldn't register collector: %s", err)))

		if err != nil {
			log.Fatal(num, err)
		}
	}

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		})
	h.ServeHTTP(w, r)
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address",
			"Address on which to expose metrics and web interface.").Default(":9356").String()
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").Default("/metrics").String()
		htmlPath = kingpin.Flag("web.html-path",
			"Path under which to expose crm_mon html output.").Default("/html").String()
		xmlPath = kingpin.Flag("web.xml-path",
			"Path under which to expose crm_mon xml output.").Default("/xml").String()
		num int
		err error
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("pacemaker_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting pacemaker_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	// This instance is only used to check collector creation and logging.
	nc, err := collector.NewPacemakerCollector()
	if err != nil {
		log.Fatalf("Couldn't create collector: %s", err)
	}

	log.Infof("Enabled collectors:")

	for n := range nc.Collectors {
		log.Infof(" - %s", n)
	}

	http.HandleFunc(*metricsPath, handler)
	http.HandleFunc("/html", collector.HTMLHandler)
	http.HandleFunc("/xml", collector.XMLHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		num, err = w.Write([]byte(`<html>
			<head><title>Pacemaker Exporter</title></head>
			<body>
			<h1>Pacemaker Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<p><a href="` + *htmlPath + `">HTML</a></p>
			<p><a href="` + *xmlPath + `">XML</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Fatal(num, err)
		}
	})

	log.Infoln("Listening on", *listenAddress)

	err = http.ListenAndServe(*listenAddress, nil)

	if err != nil {
		log.Fatal(err)
	}
}
