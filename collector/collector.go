// MIT License

// Copyright (c) 2023 Yann Vigara, Angarium Limited

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
package collector

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	binrpc "github.com/florentchauveau/go-kamailio-binrpc/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter namespace.
const namespace = "kamailio"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

const (
	defaultEnabled  = true
	defaultDisabled = false
)

var (
	factories              = make(map[string]func(config *KamailioCollectorConfig, logger log.Logger) (Collector, error))
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
	collectorStateGlobal   = make(map[string]bool)
	availableCollectors    = make([]string, 0)
)

func registerCollector(collector string, isDefaultEnabled bool, factory func(config *KamailioCollectorConfig, logger log.Logger) (Collector, error)) {
	availableCollectors = append(availableCollectors, collector)
	collectorStateGlobal[collector] = isDefaultEnabled
	factories[collector] = factory
}

// KamailioCollector implements the prometheus.Collector interface.
type KamailioCollector struct {
	Collectors map[string]Collector
	timeout    time.Duration
	url        *url.URL
	logger     log.Logger
}

// NewKamailioCollector creates a new NodeCollector.
func NewKamailioCollector(config *KamailioCollectorConfig, logger log.Logger) (*KamailioCollector, error) {
	// fill the Collector struct
	url, err := url.Parse(*config.BinrpcURI)
	if err != nil {
		return nil, fmt.Errorf("cannot parse URI: %w", err)
	}

	collectors := make(map[string]Collector)

	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()
	for key, enabled := range collectorStateGlobal {
		if !enabled {
			continue
		}
		collector, err := factories[key](config, log.With(logger, "collector", key))
		if err != nil {
			return nil, err
		}
		collectors[key] = collector
		initiatedCollectors[key] = collector
	}
	return &KamailioCollector{Collectors: collectors, logger: logger, url: url, timeout: *config.Timeout}, nil
}

// Describe implements the prometheus.Collector interface.
func (n KamailioCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (n KamailioCollector) Collect(ch chan<- prometheus.Metric) {
	var err error
	var conn net.Conn
	address := n.url.Host

	if n.url.Scheme == "unix" {
		address = n.url.Path
	}

	conn, err = net.DialTimeout(n.url.Scheme, address, n.timeout)
	if err != nil {
		level.Error(n.logger).Log("msg", "Can not connect to kamailio", "err", err)
		return
	}

	conn.SetDeadline(time.Now().Add(n.timeout))
	defer conn.Close()

	for name, c := range n.Collectors {
		execute(name, c, conn, ch, n.logger)
	}
}

func execute(name string, c Collector, conn net.Conn, ch chan<- prometheus.Metric, logger log.Logger) {
	begin := time.Now()
	err := c.Update(conn, ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			level.Debug(logger).Log("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		level.Debug(logger).Log("msg", "collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(conn net.Conn, ch chan<- prometheus.Metric) error
}

func getRecords(conn net.Conn, logger log.Logger, values ...string) ([]binrpc.Record, error) {
	cookie, err := binrpc.WritePacket(conn, values...)
	if err != nil {
		level.Error(logger).Log("msg", "Can not request", "cmd", values[0], "err", err)
		return nil, err
	}

	records, err := binrpc.ReadPacket(conn, cookie)
	if err != nil || len(records) == 0 {
		level.Error(logger).Log("msg", "Can not fetch", "cmd", values[0], "err", err)
		return nil, err
	}
	return records, nil
}
