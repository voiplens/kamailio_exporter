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
	registerCollector("tls.info", defaultEnabled, NewTLSInfoCollector)
}

type TLSInfoCollector struct {
	openedConnections *prometheus.Desc
	maxConnections    *prometheus.Desc
	clearTextWrite    *prometheus.Desc
	logger            log.Logger
	config            *KamailioCollectorConfig
}

// NewTLSInfoCollector returns a new Collector exposing TLS metrics.
func NewTLSInfoCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &TLSInfoCollector{
		openedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tls", "opened_connections"),
			"TLS Opened Connections",
			[]string{}, nil),
		maxConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tls", "max_connections"),
			"TLS Opened Connections",
			[]string{}, nil),
		clearTextWrite: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tls", "clear_text_write_queued_bytes"),
			"TLS Opened Connections",
			[]string{}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *TLSInfoCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "tls.info")
	if err != nil {
		return err
	}

	for _, record := range records {
		items, _ := record.StructItems()
		var maxConnections, openedConnections, clearTextWriteQueuedBytes int
		for _, item := range items {
			switch item.Key {
			case "max_connections":
				maxConnections, _ = item.Value.Int()
			case "opened_connections":
				openedConnections, _ = item.Value.Int()
			case "clear_text_write_queued_bytes":
				clearTextWriteQueuedBytes, _ = item.Value.Int()
			}
		}
		metricChannel <- prometheus.MustNewConstMetric(
			c.openedConnections,
			prometheus.GaugeValue,
			float64(openedConnections),
		)
		metricChannel <- prometheus.MustNewConstMetric(
			c.maxConnections,
			prometheus.GaugeValue,
			float64(maxConnections),
		)
		metricChannel <- prometheus.MustNewConstMetric(
			c.clearTextWrite,
			prometheus.GaugeValue,
			float64(clearTextWriteQueuedBytes),
		)
	}
	return nil
}
