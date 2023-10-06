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
	"strconv"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("pkg.stats", defaultEnabled, NewPkgStatsCollector)
}

type PkgStatsEntry struct {
	entry      int
	used       int
	free       int
	realUsed   int
	totalSize  int
	totalFrags int
}

type pkgStatsCollector struct {
	used   *prometheus.Desc
	free   *prometheus.Desc
	real   *prometheus.Desc
	size   *prometheus.Desc
	frags  *prometheus.Desc
	logger log.Logger
	config *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewPkgStatsCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &pkgStatsCollector{
		used: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pkgmem_used"),
			"Private memory used",
			[]string{"entry"},
			nil),

		free: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pkgmem_free"),
			"Private memory free",
			[]string{"entry"},
			nil),

		real: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pkgmem_real"),
			"Private memory real used",
			[]string{"entry"},
			nil),

		size: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pkgmem_size"),
			"Private memory total size",
			[]string{"entry"},
			nil),

		frags: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pkgmem_frags"),
			"Private memory total frags",
			[]string{"entry"},
			nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *pkgStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "pkg.stats")
	if err != nil {
		return err
	}

	// convert each pkg entry to a series of metrics
	for _, record := range records {
		items, _ := record.StructItems()
		entry := PkgStatsEntry{}
		for _, item := range items {
			switch item.Key {
			case "entry":
				entry.entry, _ = item.Value.Int()
			case "used":
				entry.used, _ = item.Value.Int()
			case "free":
				entry.free, _ = item.Value.Int()
			case "real_used":
				entry.realUsed, _ = item.Value.Int()
			case "total_size":
				entry.totalSize, _ = item.Value.Int()
			case "total_frags":
				entry.totalFrags, _ = item.Value.Int()
			}
		}
		sentry := strconv.Itoa(entry.entry)
		metricChannel <- prometheus.MustNewConstMetric(c.used, prometheus.GaugeValue, float64(entry.used), sentry)
		metricChannel <- prometheus.MustNewConstMetric(c.free, prometheus.GaugeValue, float64(entry.free), sentry)
		metricChannel <- prometheus.MustNewConstMetric(c.real, prometheus.GaugeValue, float64(entry.realUsed), sentry)
		metricChannel <- prometheus.MustNewConstMetric(c.size, prometheus.GaugeValue, float64(entry.totalSize), sentry)
		metricChannel <- prometheus.MustNewConstMetric(c.frags, prometheus.GaugeValue, float64(entry.totalFrags), sentry)
	}
	return nil
}
