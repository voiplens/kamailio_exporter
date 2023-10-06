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
	registerCollector("dlg.profile_get_size", defaultEnabled, NewDlgProfileCollector)
}

type dlgProfileCollector struct {
	logger log.Logger
	dialog *prometheus.Desc
	config *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewDlgProfileCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &dlgProfileCollector{
		dialog: prometheus.NewDesc(prometheus.BuildFQName(namespace, "dlg_profile_get_size", "dialog"), "Current number of dialogs belonging to a profile.", []string{"profile"}, nil),
		config: config,
		logger: logger,
	}, nil
}

func (c *dlgProfileCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	for _, p := range *c.config.DialogProfile.Profiles {
		records, err := getRecords(conn, c.logger, "dlg.profile_get_size", p)
		if err != nil {
			return err
		}
		i, _ := records[0].Int()
		metricChannel <- prometheus.MustNewConstMetric(c.dialog, prometheus.GaugeValue, float64(i), p)
	}

	return nil
}
