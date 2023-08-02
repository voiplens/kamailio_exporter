// MIT License

// Copyright (c) 2021 Thomas Weber, pascom GmbH & Co. Kg

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/angarium-cloud/kamailio_exporter/collector"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

func init() {
	prometheus.MustRegister(version.NewCollector("kamailio_exporter"))
}

var Version string

func main() {
	var (
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		rtpmetricsPath = kingpin.Flag(
			"web.rtp-telemetry-path",
			"Path under which to expose rtpengine metrics.",
		).Default("").String()
		socketPath = kingpin.Flag(
			"kamailio.socket-path",
			"Path to Kamailio unix domain socket",
		).Default("/var/run/kamailio/kamailio_ctl").String()
		host = kingpin.Flag(
			"kamailio.rpc-host",
			`URI of Kamailio RPC endpoint. Example: "tcp://localhost:3012"`,
		).Default("").String()
		port = kingpin.Flag(
			"kamailio.rpc-port",
			`URI of Kamailio RPC endpoint. Example: "tcp://localhost:3012"`,
		).Default("3012").Int()
		customMetricsURL = kingpin.Flag(
			"kamailio.custom-metrics-url",
			"URL to request user defined metrics from kamailio",
		).Default("").String()
		// rpcURI = kingpin.Flag(
		// 	"kamailio.rpc-uri",
		// 	`URI of Kamailio RPC endpoint. Example: "tcp://localhost:3012"`,
		// ).Default("").String()
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9494")
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("freeswitch_exporter"))
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting freeswitch_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	c, err := collector.New(*socketPath, *host, *port, logger)

	if err != nil {
		panic(err)
	}

	prometheus.MustRegister(c)

	http.Handle(*metricsPath, promhttp.Handler())
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "FreeSWITCH Exporter",
			Description: "Prometheus Exporter for FreeSWITCH servers",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}
	if *rtpmetricsPath != "" {
		level.Info(logger).Log("msg", "Enabling rtp metrics", "path", rtpmetricsPath)
		http.HandleFunc(*rtpmetricsPath, func(w http.ResponseWriter, r *http.Request) {
			resp, err := http.Get("http://127.0.0.1:9901/metrics")
			if err != nil {
				level.Warn(logger).Log("err", err)
				http.Error(w,
					fmt.Sprintf("Failed to connect to rtpengine: %s", err.Error()),
					http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()
			resp2, err := io.ReadAll(resp.Body)
			if err != nil {
				level.Warn(logger).Log("err", err)
				http.Error(w,
					fmt.Sprintf("Failed to read response from rtpengine: %s", err.Error()),
					http.StatusInternalServerError)
				return
			}
			w.Write(resp2)
		})
	}

	if *customMetricsURL != "" {
		http.Handle(*metricsPath, handlerWithUserDefinedMetrics(*customMetricsURL, logger))
	} else {
		http.Handle(*metricsPath, promhttp.Handler())
	}
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		level.Info(logger).Log("err", err)
		os.Exit(1)
	}
}

// Request user defined metrics and parse them into proper data objects
func gatherUserDefinedMetrics(url string, logger log.Logger) ([]*dto.MetricFamily, error) {
	resp, err := http.Get(url)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to query kamailio user defined metrics", "err", err)
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		level.Error(logger).Log("msg", "Requesting user defined kamailio metrics returned status code", "status", resp.StatusCode)
		return nil, err
	}

	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to read kamailio user defined metrics", "err", err)
		return nil, err
	}

	parser := expfmt.TextParser{}
	parsed, err := parser.TextToMetricFamilies(bytes.NewReader(respBytes))
	if err != nil {
		return nil, err
	}

	result := []*dto.MetricFamily{}
	for _, mf := range parsed {
		result = append(result, mf)
	}

	return result, nil
}

func handlerWithUserDefinedMetrics(userDefinedMetricsURL string, logger log.Logger) http.Handler {
	gatherer := func() ([]*dto.MetricFamily, error) {
		ours, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			return ours, err
		}
		theirs, err := gatherUserDefinedMetrics(userDefinedMetricsURL, logger)
		if err != nil {
			level.Error(logger).Log("msg", "Scraping user defined metrics failed", "err", err)
			return ours, nil
		}
		return append(ours, theirs...), nil
	}

	// defaults like promhttp.Handler(), except using our own gatherer
	return promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.GathererFunc(gatherer), promhttp.HandlerOpts{}))
}
