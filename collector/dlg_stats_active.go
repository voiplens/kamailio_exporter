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
	registerCollector("dlg.stats_active", defaultEnabled, NewDlgStatsActiveCollector)
}

type dlgStatsActiveCollector struct {
	logger log.Logger
	gauges map[string]*prometheus.Desc
	config *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewDlgStatsActiveCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	gauges := map[string]*prometheus.Desc{
		"starting":   prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_stats_active", "starting"), "Dialog starting.", []string{}, nil),
		"connecting": prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_stats_active", "connecting"), "Dialog connecting.", []string{}, nil),
		"answering":  prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_stats_active", "answering"), "Dialog answering.", []string{}, nil),
		"ongoing":    prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_stats_active", "ongoing"), "Dialog ongoing.", []string{}, nil),
		"all":        prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_stats_active", "all"), "Dialog all.", []string{}, nil),
	}
	return &dlgStatsActiveCollector{
		gauges: gauges,
		config: config,
		logger: logger,
	}, nil
}

func (c *dlgStatsActiveCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "sl.stats")
	if err != nil {
		return err
	}

	// convert each pkg entry to a series of metrics
	for _, record := range records {
		items, _ := record.StructItems()
		for _, item := range items {
			i, _ := item.Value.Int()
			if desc, ok := c.gauges[item.Key]; ok {
				metricChannel <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(i))
			}
		}
	}
	return nil
}
