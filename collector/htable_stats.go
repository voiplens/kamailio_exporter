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
	registerCollector("htable.stats", defaultEnabled, NewHtableStatsCollector)
}

type HtableStatsCollector struct {
	htableSlot  *prometheus.Desc
	htableTotal *prometheus.Desc
	htableMin   *prometheus.Desc
	htableMax   *prometheus.Desc
	logger      log.Logger
	config      *KamailioCollectorConfig
}

// NewHtableStatsCollector returns a new Collector exposing htable stats.
func NewHtableStatsCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &HtableStatsCollector{
		htableSlot: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "slots_total"),
			"Number of slots in the htable",
			[]string{"name"}, nil),
		htableTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "items_total"),
			"Total number of items stored in the htable",
			[]string{"name"}, nil),
		htableMin: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "items_per_slots_min"),
			"Min number of items per slot in the htable",
			[]string{"name"}, nil),
		htableMax: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "items_per_slots_max"),
			"Max number of items per slot in the htable",
			[]string{"name"}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *HtableStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "htable.stats")
	if err != nil {
		return err
	}

	for _, record := range records {
		items, _ := record.StructItems()
		var all, min, max, slots int
		var name string
		for _, item := range items {
			switch item.Key {
			case "slots":
				slots, _ = item.Value.Int()
			case "all":
				all, _ = item.Value.Int()
			case "min":
				min, _ = item.Value.Int()
			case "max":
				max, _ = item.Value.Int()
			case "name":
				name, _ = item.Value.String()
			}
		}
		metricChannel <- prometheus.MustNewConstMetric(c.htableSlot, prometheus.GaugeValue, float64(slots), name)
		metricChannel <- prometheus.MustNewConstMetric(c.htableTotal, prometheus.GaugeValue, float64(all), name)
		metricChannel <- prometheus.MustNewConstMetric(c.htableMin, prometheus.GaugeValue, float64(min), name)
		metricChannel <- prometheus.MustNewConstMetric(c.htableMax, prometheus.GaugeValue, float64(max), name)
	}
	return nil
}
