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
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("rtpengine.show", defaultEnabled, NewRtpengineCollector)
}

type rtpengineStatsCollector struct {
	rtpengineEnabled *prometheus.Desc
	logger           log.Logger
	config           *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewRtpengineCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &rtpengineStatsCollector{
		rtpengineEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rtpengine_enabled"),
			"rtpengine connection status",
			[]string{"url", "set", "index", "weight"},
			nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *rtpengineStatsCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	// fetch rtpengine disabled status and url
	records, err := getRecords(conn, c.logger, "rtpengine.show", "all")
	if err != nil {
		return err
	}

	for _, record := range records {
		items, _ := record.StructItems()
		if len(items) == 0 {
			level.Debug(c.logger).Log("msg", "Rtpengine.show all has empty items in record - probably because rtpengine is disabled")
			continue
		}
		var url string
		var setInt, indexInt, weightInt int
		var set, index, weight string
		var v int
		for _, item := range items {
			switch item.Key {
			case "disabled":
				v, _ = item.Value.Int()
			case "url":
				url, _ = item.Value.String()
			case "set":
				setInt, _ = item.Value.Int()
				set = strconv.Itoa(setInt)
			case "index":
				indexInt, _ = item.Value.Int()
				index = strconv.Itoa(indexInt)
			case "weight":
				weightInt, _ = item.Value.Int()
				weight = strconv.Itoa(weightInt)
			}
		}
		if url == "" {
			level.Error(c.logger).Log("msg", "No valid url found for rtpengine, failed to construct metric rtpengine_enabled")
			continue
		}
		// invert the disabled status to fit the metric name "rtpengine_enabled"
		if v == 1 {
			v = 0
		} else {
			v = 1
		}
		metricChannel <- prometheus.MustNewConstMetric(c.rtpengineEnabled, prometheus.GaugeValue, float64(v), url, set, index, weight)
	}
	return nil
}
