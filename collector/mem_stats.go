// MIT License

// Copyright (c) 2025 Alexander Bakker

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
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"go.voiplens.io/kamailio/binrpc"
)

func init() {
	registerCollector("mod.mem_stats", defaultEnabled, NewMemStatsCollector)
}

type memStatsCollector struct {
	modBytes *prometheus.Desc
	logger   log.Logger
	config   *KamailioCollectorConfig
}

// NewMemStatsCollector returns a new Collector exposing memory stats.
func NewMemStatsCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &memStatsCollector{
		modBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "shm_module_bytes"),
			"Shared memory used per module",
			[]string{"module"},
			nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *memStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "mod.mem_stats", "all", "shm")
	if err != nil {
		return err
	}

	var module string
	for _, record := range records {
		if record.Type == binrpc.TypeString {
			s, err := record.String()
			if err != nil {
				return err
			}
			const prefix = "Module: "
			if strings.HasPrefix(s, prefix) {
				module = s[len(prefix):]
			}
			continue
		}

		if record.Type == binrpc.TypeStruct {
			var total int
			items, _ := record.StructItems()
			for _, item := range items {
				switch item.Key {
				case "Total":
					total, _ = item.Value.Int()
				}
			}

			metricChannel <- prometheus.MustNewConstMetric(c.modBytes, prometheus.GaugeValue, float64(total), module)
		}
	}

	return nil
}
