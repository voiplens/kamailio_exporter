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
	"regexp"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("tm.stats", defaultEnabled, NewTmStatsCollector)
}

// this is used to match codes returned by Kamailio
// examples: "200" or "6xx" or even "xxx"
var codeRegex = regexp.MustCompile("^[0-9x]{3}$")

type tmStatsCollector struct {
	codes    *prometheus.Desc
	counters map[string]*prometheus.Desc
	gauges   map[string]*prometheus.Desc

	logger log.Logger
	config *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewTmStatsCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	gauges := map[string]*prometheus.Desc{
		"current": prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "current"), "Current transactions.", []string{}, nil),
		"waiting": prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "waiting"), "Waiting transactions.", []string{}, nil),
	}

	counters := map[string]*prometheus.Desc{
		"total":         prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "total"), "Total transactions.", []string{}, nil),
		"total_local":   prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "local_total"), "Total local transactions.", []string{}, nil),
		"rpl_received":  prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "rpl_received_total"), "Number of reply received.", []string{}, nil),
		"rpl_generated": prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "rpl_generated_total"), "Number of reply generated.", []string{}, nil),
		"rpl_sent":      prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "rpl_sent_total"), "Number of reply sent.", []string{}, nil),
		"created":       prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "created_total"), "Created transactions.", []string{}, nil),
		"freed":         prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "freed_total"), "Freed transactions.", []string{}, nil),
		"delayed_free":  prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "delayed_free_total"), "Delayed free transactions.", []string{}, nil),
	}

	return &tmStatsCollector{
		codes:    prometheus.NewDesc(prometheus.BuildFQName(namespace, "tm_stats", "codes_total"), "Per-code counters.", []string{"code"}, nil),
		counters: counters,
		gauges:   gauges,
		config:   config,
		logger:   logger,
	}, nil
}

func (c *tmStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "tm.stats")
	if err != nil {
		return err
	}

	// convert each pkg entry to a series of metrics
	for _, record := range records {
		items, _ := record.StructItems()
		for _, item := range items {
			i, _ := item.Value.Int()

			if codeRegex.MatchString(item.Key) {
				metricChannel <- prometheus.MustNewConstMetric(c.codes, prometheus.CounterValue, float64(i), item.Key)
				continue
			}

			if desc, ok := c.counters[item.Key]; ok {
				metricChannel <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(i))
				continue
			}

			if desc, ok := c.gauges[item.Key]; ok {
				metricChannel <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(i))
				continue
			}
		}
	}
	return nil
}
