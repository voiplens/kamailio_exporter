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
	registerCollector("sl.stats", defaultEnabled, NewSlStatsCollector)
}

type slStatsCollector struct {
	codes  *prometheus.Desc
	logger log.Logger
	config *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewSlStatsCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &slStatsCollector{
		codes:  prometheus.NewDesc(prometheus.BuildFQName(namespace, "sl_stats", "codes_total"), "Per-code counters.", []string{"code"}, nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *slStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "sl.stats")
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
			}
		}
	}
	return nil
}
