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
	registerCollector("htable.listTables", defaultEnabled, NewHtableListTablesCollector)
}

type HtableListTablesCollector struct {
	htableAutoExpire   *prometheus.Desc
	htableUpdateExpire *prometheus.Desc
	htableDmqReplicate *prometheus.Desc
	htableDBMode       *prometheus.Desc
	logger             log.Logger
	config             *KamailioCollectorConfig
}

// NewHtableListTablesCollector returns a new Collector exposing htables status stats.
func NewHtableListTablesCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &HtableListTablesCollector{
		htableAutoExpire: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "auto_expire_seconds"),
			"Time in seconds to delete an item from a hash table if no update was done to it",
			[]string{"name"}, nil),
		htableUpdateExpire: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "update_expire_status"),
			"Update Expire status",
			[]string{"name"}, nil),
		htableDmqReplicate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "dmq_replicate_status"),
			"DMQ Replicate status",
			[]string{"name"}, nil),
		htableDBMode: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "htable", "db_mode_status"),
			"Htable write back to db table",
			[]string{"name", "dbtable"}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *HtableListTablesCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "htable.listTables")
	if err != nil {
		return err
	}

	for _, record := range records {
		items, _ := record.StructItems()
		var dbmode, expire, updateexpire, dmqreplicate int
		var name, dbtable string
		for _, item := range items {
			switch item.Key {
			case "name":
				name, _ = item.Value.String()
			case "dbtable":
				dbtable, _ = item.Value.String()
			case "dbmode":
				dbmode, _ = item.Value.Int()
			case "expire":
				expire, _ = item.Value.Int()
			case "updateexpire":
				updateexpire, _ = item.Value.Int()
			case "dmqreplicate":
				dmqreplicate, _ = item.Value.Int()
			}
		}
		metricChannel <- prometheus.MustNewConstMetric(c.htableAutoExpire, prometheus.GaugeValue, float64(expire), name)
		metricChannel <- prometheus.MustNewConstMetric(c.htableDBMode, prometheus.GaugeValue, float64(dbmode), name, dbtable)
		metricChannel <- prometheus.MustNewConstMetric(c.htableDmqReplicate, prometheus.GaugeValue, float64(dmqreplicate), name)
		metricChannel <- prometheus.MustNewConstMetric(c.htableUpdateExpire, prometheus.GaugeValue, float64(updateexpire), name)
	}
	return nil
}
