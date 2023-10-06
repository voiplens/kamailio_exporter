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
	"net"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("core.tcp_info", defaultEnabled, NewCoreTCPInfoCollector)
}

type coreTCPInfoCollector struct {
	tcpReaders        *prometheus.Desc
	tcpMaxConnections *prometheus.Desc
	tlsMaxConnections *prometheus.Desc
	tlsConnections    *prometheus.Desc
	logger            log.Logger
	config            *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewCoreTCPInfoCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &coreTCPInfoCollector{
		tcpReaders: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tcp_readers"),
			"TCP readers",
			[]string{}, nil),
		tcpMaxConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tcp_max_connections"),
			"TCP connection limit",
			[]string{}, nil),
		tlsMaxConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tls_max_connections"),
			"TLS connection limit",
			[]string{}, nil),
		tlsConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tls_connections"),
			"Opened TLS connections",
			[]string{}, nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *coreTCPInfoCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	// fetch tcp details
	records, err := getRecords(conn, c.logger, "core.tcp_info")
	if err != nil {
		return err
	}

	items, _ := records[0].StructItems()
	var v int
	for _, item := range items {
		switch item.Key {
		case "readers":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(c.tcpReaders, prometheus.GaugeValue, float64(v))
		case "max_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(c.tcpMaxConnections, prometheus.GaugeValue, float64(v))
		case "max_tls_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(c.tlsMaxConnections, prometheus.GaugeValue, float64(v))
		case "opened_tls_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(c.tlsConnections, prometheus.GaugeValue, float64(v))
		}
	}
	return nil
}
