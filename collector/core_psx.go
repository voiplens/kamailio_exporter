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
	registerCollector("core.psa", defaultEnabled, NewCorePsaCollector)
}

type CorePsxCollector struct {
	coreProcessStatus *prometheus.Desc
	logger            log.Logger
	config            *KamailioCollectorConfig
}

// NewDispatcherListCollector returns a new Collector exposing core processes stats.
func NewCorePsaCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &CorePsxCollector{
		coreProcessStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_process_status"),
			"Status of each process running in Kamailio",
			[]string{"index", "pid", "rank", "description"}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *CorePsxCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "core.psa")
	if err != nil {
		return err
	}

	for _, record := range records {
		items, _ := record.StructItems()
		var index, pid, rank, status int
		var description string
		for _, item := range items {
			switch item.Key {
			case "index":
				index, _ = item.Value.Int()
			case "pid":
				pid, _ = item.Value.Int()
			case "status":
				status, _ = item.Value.Int()
			case "rank":
				rank, _ = item.Value.Int()
			case "description":
				description, _ = item.Value.String()
			}
		}
		metricChannel <- prometheus.MustNewConstMetric(
			c.coreProcessStatus,
			prometheus.GaugeValue,
			float64(status),
			strconv.Itoa(index),
			strconv.Itoa(pid),
			strconv.Itoa(rank),
			description,
		)
	}
	return nil
}
