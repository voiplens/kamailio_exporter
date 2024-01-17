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
	registerCollector("core.runinfo", defaultEnabled, NewCoreRuninfoCollector)
}

type CoreRuninfoCollector struct {
	coreUptime *prometheus.Desc
	logger     log.Logger
	config     *KamailioCollectorConfig
}

// NewStatsFetchCollector returns a new Collector exposing core stats.
func NewCoreRuninfoCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &CoreRuninfoCollector{
		coreUptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_uptime"),
			"Uptime in seconds",
			[]string{"version", "compiled", "compiler"}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *CoreRuninfoCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "core.runinfo")
	if err != nil {
		return err
	}

	items, _ := records[0].StructItems()
	var version string
	var compiled, compiler string
	var uptime int
	for _, item := range items {
		switch item.Key {
		case "version":
			version, _ = item.Value.String()
		case "compiled":
			compiled, _ = item.Value.String()
		case "compiler":
			compiler, _ = item.Value.String()
		case "uptime_secs":
			uptime, _ = item.Value.Int()
		}
	}
	metricChannel <- prometheus.MustNewConstMetric(c.coreUptime, prometheus.GaugeValue, float64(uptime), version, compiled, compiler)
	return nil
}
