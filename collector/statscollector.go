// MIT License

// Copyright (c) 2021 Thomas Weber, pascom GmbH & Co. Kg

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
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	binrpc "github.com/florentchauveau/go-kamailio-binrpc/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// declare a series of prometheus metric descriptions
// we can reuse them for each scrape
var (
	coreUptime = prometheus.NewDesc(
		"kamailio_core_uptime",
		"Uptime in seconds",
		[]string{"version", "compiled", "compiler"}, nil)

	coreRequestTotal = prometheus.NewDesc(
		"kamailio_core_request_total",
		"Request counters",
		[]string{"method"}, nil)

	coreRcvRequestTotal = prometheus.NewDesc(
		"kamailio_core_rcv_request_total",
		"Received requests by method",
		[]string{"method"}, nil)

	coreReplyTotal = prometheus.NewDesc(
		"kamailio_core_reply_total",
		"Reply counters",
		[]string{"type"}, nil)

	coreRcvReplyTotal = prometheus.NewDesc(
		"kamailio_core_rcv_reply_total",
		"Received replies by code",
		[]string{"code"}, nil)

	shmemBytes = prometheus.NewDesc(
		"kamailio_shm_bytes",
		"Shared memory sizes",
		[]string{"type"}, nil)

	shmemFragments = prometheus.NewDesc(
		"kamailio_shm_fragments",
		"Shared memory fragment count",
		[]string{}, nil)

	dnsFailed = prometheus.NewDesc(
		"kamailio_dns_failed_request_total",
		"Failed dns requests",
		[]string{}, nil)

	badURI = prometheus.NewDesc(
		"kamailio_bad_uri_total",
		"Messages with bad uri",
		[]string{}, nil)

	badMsgHdr = prometheus.NewDesc(
		"kamailio_bad_msg_hdr",
		"Messages with bad message header",
		[]string{}, nil)

	slReplyTotal = prometheus.NewDesc(
		"kamailio_sl_reply_total",
		"Stateless replies by code",
		[]string{"code"}, nil)

	slTypeTotal = prometheus.NewDesc(
		"kamailio_sl_type_total",
		"Stateless replies by type",
		[]string{"type"}, nil)

	tcpTotal = prometheus.NewDesc(
		"kamailio_tcp_total",
		"TCP connection counters",
		[]string{"type"}, nil)

	tcpConnections = prometheus.NewDesc(
		"kamailio_tcp_connections",
		"Opened TCP connections",
		[]string{}, nil)

	tcpWritequeue = prometheus.NewDesc(
		"kamailio_tcp_writequeue",
		"TCP write queue size",
		[]string{}, nil)

	tmxCodeTotal = prometheus.NewDesc(
		"kamailio_tmx_code_total",
		"Completed Transaction counters by code",
		[]string{"code"}, nil)

	tmxTypeTotal = prometheus.NewDesc(
		"kamailio_tmx_type_total",
		"Completed Transaction counters by type",
		[]string{"type"}, nil)

	tmx = prometheus.NewDesc(
		"kamailio_tmx",
		"Ongoing Transactions",
		[]string{"type"}, nil)

	tmxRplTotal = prometheus.NewDesc(
		"kamailio_tmx_rpl_total",
		"Tmx reply counters",
		[]string{"type"}, nil)

	dialog = prometheus.NewDesc(
		"kamailio_dialog",
		"Ongoing Dialogs",
		[]string{"type"}, nil)

	pkgmemUsed = prometheus.NewDesc(
		"kamailio_pkgmem_used",
		"Private memory used",
		[]string{"entry"},
		nil)

	pkgmemFree = prometheus.NewDesc(
		"kamailio_pkgmem_free",
		"Private memory free",
		[]string{"entry"},
		nil)

	pkgmemReal = prometheus.NewDesc(
		"kamailio_pkgmem_real",
		"Private memory real used",
		[]string{"entry"},
		nil)

	pkgmemSize = prometheus.NewDesc(
		"kamailio_pkgmem_size",
		"Private memory total size",
		[]string{"entry"},
		nil)

	pkgmemFrags = prometheus.NewDesc(
		"kamailio_pkgmem_frags",
		"Private memory total frags",
		[]string{"entry"},
		nil)

	tcpReaders = prometheus.NewDesc(
		"kamailio_tcp_readers",
		"TCP readers",
		[]string{},
		nil)

	tcpMaxConnections = prometheus.NewDesc(
		"kamailio_tcp_max_connections",
		"TCP connection limit",
		[]string{},
		nil)

	tlsMaxConnections = prometheus.NewDesc(
		"kamailio_tls_max_connections",
		"TLS connection limit",
		[]string{},
		nil)

	tlsConnections = prometheus.NewDesc(
		"kamailio_tls_connections",
		"Opened TLS connections",
		[]string{},
		nil)

	rtpengineEnabled = prometheus.NewDesc(
		"kamailio_rtpengine_enabled",
		"rtpengine connection status",
		[]string{"url", "set", "index", "weight"},
		nil)
)

type PkgStatsEntry struct {
	entry      int
	used       int
	free       int
	realUsed   int
	totalSize  int
	totalFrags int
}

// the actual Collector object
type StatsCollector struct {
	binrpcURI string
	timeout   time.Duration

	url    *url.URL
	logger log.Logger
}

// produce a new StatsCollector object
func New(binrpcURI string, timeout time.Duration, logger log.Logger) (*StatsCollector, error) {
	// fill the Collector struct
	var c StatsCollector
	c.logger = logger
	c.timeout = timeout
	c.binrpcURI = binrpcURI

	var url *url.URL
	var err error

	if url, err = url.Parse(c.binrpcURI); err != nil {
		return nil, fmt.Errorf("cannot parse URI: %w", err)
	}

	c.url = url

	// fine, return the created object struct
	return &c, nil
}

// part of the prometheus.Collector interface
func (c *StatsCollector) Describe(descriptionChannel chan<- *prometheus.Desc) {
	// DescribeByCollect is a helper to implement the Describe method of a custom
	// Collector. It collects the metrics from the provided Collector and sends
	// their descriptors to the provided channel.
	prometheus.DescribeByCollect(c, descriptionChannel)
}

// part of the prometheus.Collector interface
func (c *StatsCollector) Collect(metricChannel chan<- prometheus.Metric) {
	// TODO measure rpc time
	//timer := prometheus.NewTimer(rpc_request_duration)
	//defer timer.ObserveDuration()

	// establish connection to Kamailio server
	var err error
	var conn net.Conn
	address := c.url.Host

	if c.url.Scheme == "unix" {
		address = c.url.Path
	}

	conn, err = net.DialTimeout(c.url.Scheme, address, c.timeout)
	if err != nil {
		level.Error(c.logger).Log("msg", "Can not connect to kamailio", "err", err)
		return
	}

	conn.SetDeadline(time.Now().Add(c.timeout))
	defer conn.Close()

	err = fetchStats(conn, c, metricChannel)
	if err != nil {
		return
	}

	err = fetchPkgStats(conn, c, metricChannel)
	if err != nil {
		return
	}

	err = fetchTCPDetails(conn, c, metricChannel)
	if err != nil {
		return
	}

	err = fetchRTPEngine(conn, c, metricChannel)
	if err != nil {
		return
	}

	err = fetchCoreUptimeAndInfo(conn, c, metricChannel)
	if err != nil {
		return
	}
}

func fetchStats(conn net.Conn, c *StatsCollector, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c, "stats.fetch", "all")
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
	produceMetrics(completeStatMap, metricChannel)
	// produce prometheus.Metric objects for scripted stats (if any)
	convertScriptedMetrics(completeStatMap, metricChannel)
	return nil
}

func fetchPkgStats(conn net.Conn, c *StatsCollector, metricChannel chan<- prometheus.Metric) error {
	// now fetch pkg stats
	records, err := getRecords(conn, c, "pkg.stats")
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
		metricChannel <- prometheus.MustNewConstMetric(pkgmemUsed, prometheus.GaugeValue, float64(entry.used), sentry)
		metricChannel <- prometheus.MustNewConstMetric(pkgmemFree, prometheus.GaugeValue, float64(entry.free), sentry)
		metricChannel <- prometheus.MustNewConstMetric(pkgmemReal, prometheus.GaugeValue, float64(entry.realUsed), sentry)
		metricChannel <- prometheus.MustNewConstMetric(pkgmemSize, prometheus.GaugeValue, float64(entry.totalSize), sentry)
		metricChannel <- prometheus.MustNewConstMetric(pkgmemFrags, prometheus.GaugeValue, float64(entry.totalFrags), sentry)
	}
	return nil
}

func fetchTCPDetails(conn net.Conn, c *StatsCollector, metricChannel chan<- prometheus.Metric) error {
	// fetch tcp details
	records, err := getRecords(conn, c, "core.tcp_info")
	if err != nil {
		return err
	}

	items, _ := records[0].StructItems()
	var v int
	for _, item := range items {
		switch item.Key {
		case "readers":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(tcpReaders, prometheus.GaugeValue, float64(v))
		case "max_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(tcpMaxConnections, prometheus.GaugeValue, float64(v))
		case "max_tls_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(tlsMaxConnections, prometheus.GaugeValue, float64(v))
		case "opened_tls_connections":
			v, _ = item.Value.Int()
			metricChannel <- prometheus.MustNewConstMetric(tlsConnections, prometheus.GaugeValue, float64(v))
		}
	}
	return nil
}

func fetchRTPEngine(conn net.Conn, c *StatsCollector, metricChannel chan<- prometheus.Metric) error {
	// fetch rtpengine disabled status and url
	records, err := getRecords(conn, c, "rtpengine.show", "all")
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
		//invert the disabled status to fit the metric name "rtpengine_enabled"
		if v == 1 {
			v = 0
		} else {
			v = 1
		}
		metricChannel <- prometheus.MustNewConstMetric(rtpengineEnabled, prometheus.GaugeValue, float64(v), url, set, index, weight)
	}
	return nil
}

func fetchCoreUptimeAndInfo(conn net.Conn, c *StatsCollector, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c, "core.info")
	if err != nil {
		return err
	}

	uptime, err := fetchUptime(conn, c)
	if err != nil {
		return err
	}

	items, _ := records[0].StructItems()
	var version string
	var compiled, compiler string
	for _, item := range items {
		switch item.Key {
		case "version":
			version, _ = item.Value.String()
		case "compiled":
			compiled, _ = item.Value.String()
		case "compiler":
			compiler, _ = item.Value.String()
		}
	}
	metricChannel <- prometheus.MustNewConstMetric(coreUptime, prometheus.GaugeValue, float64(uptime), version, compiled, compiler)
	return nil
}

func fetchUptime(conn net.Conn, c *StatsCollector) (int, error) {
	records, err := getRecords(conn, c, "core.uptime")
	if err != nil {
		return 0, err
	}

	items, _ := records[0].StructItems()
	var uptime int
	for _, item := range items {
		switch item.Key {
		case "uptime":
			uptime, _ = item.Value.Int()
		}
	}
	return uptime, nil
}

func getRecords(conn net.Conn, c *StatsCollector, values ...string) ([]binrpc.Record, error) {
	cookie, err := binrpc.WritePacket(conn, values...)
	if err != nil {
		level.Error(c.logger).Log("msg", "Can not request", "cmd", values[0], "err", err)
		return nil, err
	}

	records, err := binrpc.ReadPacket(conn, cookie)
	if err != nil || len(records) == 0 {
		level.Error(c.logger).Log("msg", "Can not fetch", "cmd", values[0], "err", err)
		return nil, err
	}
	return records, nil
}

// produce a series of prometheus.Metric values by converting "well-known" prometheus stats
func produceMetrics(completeStatMap map[string]string, metricChannel chan<- prometheus.Metric) {
	// kamailio_core_request_total
	convertStatToMetric(completeStatMap, "core.drop_requests", "drop", coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.err_requests", "err", coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.fwd_requests", "fwd", coreRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests", "rcv", coreRequestTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_rcv_request_total
	convertStatToMetric(completeStatMap, "core.rcv_requests_ack", "ack", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_bye", "bye", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_cancel", "cancel", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_info", "info", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_invite", "invite", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_message", "message", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_notify", "notify", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_options", "options", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_prack", "prack", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_publish", "publish", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_refer", "refer", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_register", "register", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_subscribe", "subscribe", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_requests_update", "update", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.unsupported_methods", "unsupported", coreRcvRequestTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_reply_total
	convertStatToMetric(completeStatMap, "core.drop_replies", "drop", coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.err_replies", "err", coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.fwd_replies", "fwd", coreReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies", "rcv", coreReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_core_rcv_reply_total
	convertStatToMetric(completeStatMap, "core.rcv_replies_18x", "18x", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_1xx", "1xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_2xx", "2xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_3xx", "3xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_401", "401", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_404", "404", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_407", "407", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_408", "408", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_480", "480", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_486", "486", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_4xx", "4xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_5xx", "5xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.rcv_replies_6xx", "6xx", coreRcvReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_shm_bytes
	convertStatToMetric(completeStatMap, "shmem.free_size", "free", shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.max_used_size", "max_used", shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.real_used_size", "real_used", shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.total_size", "total", shmemBytes, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "shmem.used_size", "used", shmemBytes, metricChannel, prometheus.GaugeValue)

	convertStatToMetric(completeStatMap, "shmem.fragments", "", shmemFragments, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "dns.failed_dns_request", "", dnsFailed, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.bad_URIs_rcvd", "", badURI, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "core.bad_msg_hdr", "", badMsgHdr, metricChannel, prometheus.CounterValue)

	// kamailio_sl_reply_total
	convertStatToMetric(completeStatMap, "sl.1xx_replies", "1xx", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.200_replies", "200", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.202_replies", "202", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.2xx_replies", "2xx", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.300_replies", "300", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.301_replies", "301", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.302_replies", "302", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.3xx_replies", "3xx", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.400_replies", "400", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.401_replies", "401", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.403_replies", "403", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.404_replies", "404", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.407_replies", "407", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.408_replies", "408", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.483_replies", "483", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.4xx_replies", "4xx", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.500_replies", "500", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.5xx_replies", "5xx", slReplyTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.6xx_replies", "6xx", slReplyTotal, metricChannel, prometheus.CounterValue)

	// kamailio_sl_type_total
	convertStatToMetric(completeStatMap, "sl.failures", "failure", slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.received_ACKs", "received_ack", slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.sent_err_replies", "sent_err_reply", slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.sent_replies", "sent_reply", slTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "sl.xxx_replies", "xxx_reply", slTypeTotal, metricChannel, prometheus.CounterValue)

	// kamailio_tcp_total
	convertStatToMetric(completeStatMap, "tcp.con_reset", "con_reset", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.con_timeout", "con_timeout", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.connect_failed", "connect_failed", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.connect_success", "connect_success", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.established", "established", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.local_reject", "local_reject", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.passive_open", "passive_open", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.send_timeout", "send_timeout", tcpTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tcp.sendq_full", "sendq_full", tcpTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tcp_connections
	convertStatToMetric(completeStatMap, "tcp.current_opened_connections", "", tcpConnections, metricChannel, prometheus.GaugeValue)
	// kamailio_tcp_writequeue
	convertStatToMetric(completeStatMap, "tcp.current_write_queue_size", "", tcpWritequeue, metricChannel, prometheus.GaugeValue)

	// kamailio_tmx_code_total
	convertStatToMetric(completeStatMap, "tmx.2xx_transactions", "2xx", tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.3xx_transactions", "3xx", tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.4xx_transactions", "4xx", tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.5xx_transactions", "5xx", tmxCodeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.6xx_transactions", "6xx", tmxCodeTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tmx_type_total
	convertStatToMetric(completeStatMap, "tmx.UAC_transactions", "uac", tmxTypeTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.UAS_transactions", "uas", tmxTypeTotal, metricChannel, prometheus.CounterValue)
	// kamailio_tmx
	convertStatToMetric(completeStatMap, "tmx.active_transactions", "active", tmx, metricChannel, prometheus.GaugeValue)
	convertStatToMetric(completeStatMap, "tmx.inuse_transactions", "inuse", tmx, metricChannel, prometheus.GaugeValue)

	// kamailio_tmx_rpl_total
	convertStatToMetric(completeStatMap, "tmx.rpl_absorbed", "absorbed", tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_generated", "generated", tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_received", "received", tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_relayed", "relayed", tmxRplTotal, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "tmx.rpl_sent", "sent", tmxRplTotal, metricChannel, prometheus.CounterValue)

	// kamailio_dialog
	convertStatToMetric(completeStatMap, "dialog.active_dialogs", "active_dialogs", dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.early_dialogs", "early_dialogs", dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.expired_dialogs", "expired_dialogs", dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.failed_dialogs", "failed_dialogs", dialog, metricChannel, prometheus.CounterValue)
	convertStatToMetric(completeStatMap, "dialog.processed_dialogs", "processed_dialogs", dialog, metricChannel, prometheus.CounterValue)
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
