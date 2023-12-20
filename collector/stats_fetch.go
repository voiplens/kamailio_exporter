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
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("stats.fetch", defaultEnabled, NewStatsFetchCollector)
}

type StatsFetchCollector struct {
	coreRequestTotal    *prometheus.Desc
	coreRcvRequestTotal *prometheus.Desc
	coreReplyTotal      *prometheus.Desc
	coreRcvReplyTotal   *prometheus.Desc
	shmemBytes          *prometheus.Desc
	shmemFragments      *prometheus.Desc
	dnsFailed           *prometheus.Desc
	badURI              *prometheus.Desc
	badMsgHdr           *prometheus.Desc
	slReplyTotal        *prometheus.Desc
	slTypeTotal         *prometheus.Desc
	tcpTotal            *prometheus.Desc
	tcpConnections      *prometheus.Desc
	tcpWritequeue       *prometheus.Desc
	tmxCodeTotal        *prometheus.Desc
	tmxTypeTotal        *prometheus.Desc
	tmx                 *prometheus.Desc
	tmxRplTotal         *prometheus.Desc
	dialog              *prometheus.Desc
	logger              log.Logger
	config              *KamailioCollectorConfig
}

// NewStatsFetchCollector returns a new Collector exposing core stats.
func NewStatsFetchCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &StatsFetchCollector{
		coreRequestTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_request_total"),
			"Request counters",
			[]string{"method"}, nil),

		coreRcvRequestTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_rcv_request_total"),
			"Received requests by method",
			[]string{"method"}, nil),

		coreReplyTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_reply_total"),
			"Reply counters",
			[]string{"type"}, nil),

		coreRcvReplyTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "core_rcv_reply_total"),
			"Received replies by code",
			[]string{"code"}, nil),

		shmemBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "shm_bytes"),
			"Shared memory sizes",
			[]string{"type"}, nil),

		shmemFragments: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "shm_fragments"),
			"Shared memory fragment count",
			[]string{}, nil),

		dnsFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "dns_failed_request_total"),
			"Failed dns requests",
			[]string{}, nil),

		badURI: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bad_uri_total"),
			"Messages with bad uri",
			[]string{}, nil),

		badMsgHdr: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bad_msg_hdr"),
			"Messages with bad message header",
			[]string{}, nil),

		slReplyTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "sl_reply_total"),
			"Stateless replies by code",
			[]string{"code"}, nil),

		slTypeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "sl_type_total"),
			"Stateless replies by type",
			[]string{"type"}, nil),

		tcpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tcp_total"),
			"TCP connection counters",
			[]string{"type"}, nil),

		tcpConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tcp_connections"),
			"Opened TCP connections",
			[]string{}, nil),

		tcpWritequeue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tcp_writequeue"),
			"TCP write queue size",
			[]string{}, nil),

		tmxCodeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tmx_code_total"),
			"Completed Transaction counters by code",
			[]string{"code"}, nil),

		tmxTypeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tmx_type_total"),
			"Completed Transaction counters by type",
			[]string{"type"}, nil),

		tmx: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tmx"),
			"Ongoing Transactions",
			[]string{"type"}, nil),

		tmxRplTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "tmx_rpl_total"),
			"Tmx reply counters",
			[]string{"type"}, nil),

		dialog: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "dialog"),
			"Ongoing Dialogs",
			[]string{"type"}, nil),
		logger: logger,
		config: config,
	}, nil
}

func (c *StatsFetchCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "stats.fetch", "all")
	if err != nil {
		return err
	}

	// convert the structure into a simple key=>value map
	items, _ := records[0].StructItems()
	completeStatMap := make(map[string]string)
	for _, item := range items {
		value, _ := item.Value.String()
		completeStatMap[item.Key] = value
	}
	// and produce various prometheus.Metric for well-known stats
	produceMetrics(completeStatMap, c, metricChannel)
	// produce prometheus.Metric objects for scripted stats (if any)
	convertScriptedMetrics(completeStatMap, metricChannel)

	return nil
}

// produce a series of prometheus.Metric values by converting "well-known" prometheus stats
func produceMetrics(completeStatMap map[string]string, c *StatsFetchCollector, metricChannel chan<- prometheus.Metric) {
	// kamailio_core_request_total
	convertStatToMetric(completeStatMap, "core.drop_requests", "drop", c.coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.err_requests", "err", c.coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.fwd_requests", "fwd", c.coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests", "rcv", c.coreRequestTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_rcv_request_total
	convertStatToMetric(completeStatMap, "core.rcv_requests_ack", "ack", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_bye", "bye", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_cancel", "cancel", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_info", "info", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_invite", "invite", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_message", "message", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_notify", "notify", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_options", "options", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_prack", "prack", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_publish", "publish", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_refer", "refer", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_register", "register", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_subscribe", "subscribe", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_update", "update", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.unsupported_methods", "unsupported", c.coreRcvRequestTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_reply_total
	convertStatToMetric(completeStatMap, "core.drop_replies", "drop", c.coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.err_replies", "err", c.coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.fwd_replies", "fwd", c.coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies", "rcv", c.coreReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_rcv_reply_total
	convertStatToMetric(completeStatMap, "core.rcv_replies_18x", "18x", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_1xx", "1xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_2xx", "2xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_3xx", "3xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_401", "401", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_404", "404", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_407", "407", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_408", "408", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_480", "480", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_486", "486", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_4xx", "4xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_5xx", "5xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_6xx", "6xx", c.coreRcvReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_shm_bytes
	convertStatToMetric(completeStatMap, "shmem.free_size", "free", c.shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.max_used_size", "max_used", c.shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.real_used_size", "real_used", c.shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.total_size", "total", c.shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.used_size", "used", c.shmemBytes, metricChannel, prometheus.GaugeValue)

	convertStatToMetric(completeStatMap, "shmem.fragments", "", c.shmemFragments, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "dns.failed_dns_request", "", c.dnsFailed, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.bad_URIs_rcvd", "", c.badURI, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.bad_msg_hdr", "", c.badMsgHdr, metricChannel, prometheus.CounterValue)

	// kamailio_sl_reply_total
	convertStatToMetric(completeStatMap, "sl.1xx_replies", "1xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.200_replies", "200", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.202_replies", "202", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.2xx_replies", "2xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.300_replies", "300", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.301_replies", "301", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.302_replies", "302", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.3xx_replies", "3xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.400_replies", "400", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.401_replies", "401", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.403_replies", "403", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.404_replies", "404", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.407_replies", "407", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.408_replies", "408", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.483_replies", "483", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.4xx_replies", "4xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.500_replies", "500", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.5xx_replies", "5xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.6xx_replies", "6xx", c.slReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_sl_type_total
	convertStatToMetric(completeStatMap, "sl.failures", "failure", c.slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.received_ACKs", "received_ack", c.slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.sent_err_replies", "sent_err_reply", c.slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.sent_replies", "sent_reply", c.slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.xxx_replies", "xxx_reply", c.slTypeTotal, metricChannel, prometheus.CounterValue)

	// kamailio_tcp_total
	convertStatToMetric(completeStatMap, "tcp.con_reset", "con_reset", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.con_timeout", "con_timeout", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.connect_failed", "connect_failed", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.connect_success", "connect_success", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.established", "established", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.local_reject", "local_reject", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.passive_open", "passive_open", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.send_timeout", "send_timeout", c.tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.sendq_full", "sendq_full", c.tcpTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tcp_connections
	convertStatToMetric(completeStatMap, "tcp.current_opened_connections", "", c.tcpConnections, metricChannel, prometheus.GaugeValue)
	// kamailio_tcp_writequeue
	convertStatToMetric(completeStatMap, "tcp.current_write_queue_size", "", c.tcpWritequeue, metricChannel, prometheus.GaugeValue)

	// kamailio_tmx_code_total
	convertStatToMetric(completeStatMap, "tmx.2xx_transactions", "2xx", c.tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.3xx_transactions", "3xx", c.tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.4xx_transactions", "4xx", c.tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.5xx_transactions", "5xx", c.tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.6xx_transactions", "6xx", c.tmxCodeTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tmx_type_total
	convertStatToMetric(completeStatMap, "tmx.UAC_transactions", "uac", c.tmxTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.UAS_transactions", "uas", c.tmxTypeTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tmx
	convertStatToMetric(completeStatMap, "tmx.active_transactions", "active", c.tmx, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "tmx.inuse_transactions", "inuse", c.tmx, metricChannel, prometheus.GaugeValue)

	// kamailio_tmx_rpl_total
	convertStatToMetric(completeStatMap, "tmx.rpl_absorbed", "absorbed", c.tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_generated", "generated", c.tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_received", "received", c.tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_relayed", "relayed", c.tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_sent", "sent", c.tmxRplTotal, metricChannel, prometheus.CounterValue)

	// kamailio_dialog
	convertStatToMetric(completeStatMap, "dialog.active_dialogs", "active_dialogs", c.dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.early_dialogs", "early_dialogs", c.dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.expired_dialogs", "expired_dialogs", c.dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.failed_dialogs", "failed_dialogs", c.dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.processed_dialogs", "processed_dialogs", c.dialog, metricChannel, prometheus.CounterValue)
}

// Iterate all reported "stats" keys and find those with a prefix of "script."
// These values are user-defined and populated within the kamailio script.
// See https://www.kamailio.org/docs/modules/5.2.x/modules/statistics.html
func convertScriptedMetrics(data map[string]string, prom chan<- prometheus.Metric) {
	for k := range data {
		// k = "script.custom_total"
		if strings.HasPrefix(k, "script.") {
			// metricName = "custom_total"
			metricName := strings.TrimPrefix(k, "script.")
			metricName = strings.ToLower(metricName)
			var valueType prometheus.ValueType
			// deduce the metrics value type by following https://prometheus.io/docs/practices/naming/
			if strings.HasSuffix(k, "_total") || strings.HasSuffix(k, "_seconds") || strings.HasSuffix(k, "_bytes") {
				valueType = prometheus.CounterValue
			} else {
				valueType = prometheus.GaugeValue
			}
			// create a metric description on the fly
			description := prometheus.NewDesc("kamailio_"+metricName, "Scripted metric "+metricName, []string{}, nil)
			// and produce a metric
			convertStatToMetric(data, k, "", description, prom, valueType)
		}
	}
}

// convert a single "stat" value to a prometheus metric
// invalid "stat" paires are skipped but logged
func convertStatToMetric(completeStatMap map[string]string, statKey string, optionalLabelValue string, metricDescription *prometheus.Desc, metricChannel chan<- prometheus.Metric, valueType prometheus.ValueType) {
	// check wether we got a labelValue or not
	var labelValues []string
	if optionalLabelValue != "" {
		labelValues = []string{optionalLabelValue}
	} else {
		labelValues = []string{}
	}
	// get the stat-value ...
	if valueAsString, ok := completeStatMap[statKey]; ok {
		// ... convert it to a float
		if value, err := strconv.ParseFloat(valueAsString, 64); err == nil {
			// and produce a prometheus metric
			metric, err := prometheus.NewConstMetric(
				metricDescription,
				valueType,
				value,
				labelValues...,
			)
			if err == nil {
				// handover the metric to prometheus api
				metricChannel <- metric
				// } else {
				// or skip and complain
				// log.Warnf("Could not convert stat value [%s]: %s", statKey, err)
			}
		}
		// } else {
		// skip stat values not found in completeStatMap
		// can happen if some kamailio modules are not loaded
		// and thus certain stat entries are not created
		// log.Debugf("Skipping stat value [%s], it was not returned by kamailio", statKey)
	}
}
