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
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	binrpc "github.com/florentchauveau/go-kamailio-binrpc/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("dispatcher.list", defaultEnabled, NewDispatcherListCollector)
}

// DispatcherTarget is a target of the dispatcher module.
type DispatcherTarget struct {
	ID             int
	URI            string
	Flags          string
	Priority       int
	Status         float64
	Body           string
	Weight         int
	RWeight        int
	Socket         string
	SipTarget      string
	LatencyAvg     float64
	LatencyStd     float64
	LatencyEst     float64
	LatencyMax     float64
	LatencyTimeout float64
}

type dispatcherListCollector struct {
	logger         log.Logger
	target         *prometheus.Desc
	latencyAvg     *prometheus.Desc
	latencyStd     *prometheus.Desc
	latencyEst     *prometheus.Desc
	latencyMax     *prometheus.Desc
	latencyTimeout *prometheus.Desc
	weight         *prometheus.Desc
	rweight        *prometheus.Desc
	priority       *prometheus.Desc
	config         *KamailioCollectorConfig
}

// NewCoreStatsCollector returns a new Collector exposing core stats.
func NewDispatcherListCollector(config *KamailioCollectorConfig, logger log.Logger) (Collector, error) {
	return &dispatcherListCollector{
		config:         config,
		logger:         logger,
		target:         prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target"), "Target status.", []string{"set_id", "destination", "set_name"}, nil),
		latencyAvg:     prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_latency_avg"), "Target Latency Average.", []string{"set_id", "destination", "set_name"}, nil),
		latencyStd:     prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_latency_std"), "Target Latency.", []string{"set_id", "destination", "set_name"}, nil),
		latencyEst:     prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_latency_est"), "Target Latency.", []string{"set_id", "destination", "set_name"}, nil),
		latencyMax:     prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_latency_max"), "Target Latency.", []string{"set_id", "destination", "set_name"}, nil),
		latencyTimeout: prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_latency_timeout"), "Target Latency.", []string{"set_id", "destination", "set_name"}, nil),
		weight:         prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_weight"), "Target Weight.", []string{"set_id", "destination", "set_name"}, nil),
		rweight:        prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_rweight"), "Target rweight.", []string{"set_id", "destination", "set_name"}, nil),
		priority:       prometheus.NewDesc(prometheus.BuildFQName(namespace, "dispatcher_list", "target_priority"), "Target Priority.", []string{"set_id", "destination", "set_name"}, nil),
	}, nil
}

func (c *dispatcherListCollector) Update(conn net.Conn, metricChannel chan<- prometheus.Metric) error {
	records, err := getRecords(conn, c.logger, "dispatcher.list")
	if err != nil {
		return err
	}

	targets, err := parseDispatcherTargets(records)
	if err != nil {
		return err
	}

	// convert each pkg entry to a series of metrics
	for _, target := range targets {
		setID := fmt.Sprintf("%d", target.ID)
		setName := c.config.DispatcherMap[target.ID]
		metricChannel <- prometheus.MustNewConstMetric(c.target, prometheus.GaugeValue, target.Status, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.latencyAvg, prometheus.GaugeValue, target.LatencyAvg, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.latencyStd, prometheus.GaugeValue, target.LatencyStd, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.latencyEst, prometheus.GaugeValue, target.LatencyEst, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.latencyMax, prometheus.GaugeValue, target.LatencyMax, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.latencyTimeout, prometheus.GaugeValue, target.LatencyTimeout, setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.priority, prometheus.GaugeValue, float64(target.Priority), setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.weight, prometheus.GaugeValue, float64(target.Weight), setID, target.URI, setName)
		metricChannel <- prometheus.MustNewConstMetric(c.rweight, prometheus.GaugeValue, float64(target.RWeight), setID, target.URI, setName)
	}
	return nil
}

// parseDispatcherTargets parses the "dispatcher.list" result and returns a list of targets.
func parseDispatcherTargets(records []binrpc.Record) ([]DispatcherTarget, error) {
	var targets []DispatcherTarget
	for _, record := range records {
		items, _ := record.StructItems()

		result, err := parseRecords(items)
		if err != nil {
			return nil, err
		}
		targets = append(targets, result...)
	}
	return targets, nil
}

func parseRecords(items []binrpc.StructItem) ([]DispatcherTarget, error) {
	var targets []DispatcherTarget
	for _, item := range items {
		if item.Key != "RECORDS" {
			continue
		}

		sets, err := item.Value.StructItems()
		if err != nil {
			return nil, err
		}

		for _, item := range sets {
			if item.Key != "SET" {
				continue
			}

			setItems, err := item.Value.StructItems()
			if err != nil {
				return nil, err
			}

			result, err := parseSetItems(setItems)
			if err != nil {
				return nil, err
			}
			targets = append(targets, result...)
		}
	}
	return targets, nil
}

func parseSetItems(setItems []binrpc.StructItem) ([]DispatcherTarget, error) {
	var setID int
	var destinations []binrpc.StructItem
	var err error

	for _, set := range setItems {
		if set.Key == "ID" {
			if setID, err = set.Value.Int(); err != nil {
				return nil, err
			}
		}
		if set.Key == "TARGETS" {
			destinations, err = set.Value.StructItems()
			if err != nil {
				return nil, err
			}
		}
	}

	if setID == 0 {
		return nil, errors.New("missing set ID while parsing dispatcher.list")
	}

	targets, err := parseDestinations(setID, destinations)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

func parseDestinations(setID int, destinations []binrpc.StructItem) ([]DispatcherTarget, error) {
	var targets []DispatcherTarget
	for _, destination := range destinations {
		if destination.Key != "DEST" {
			continue
		}

		props, err := destination.Value.StructItems()

		if err != nil {
			return nil, err
		}

		target := DispatcherTarget{ID: setID}

		for _, prop := range props {
			switch prop.Key {
			case "URI":
				target.URI, err = prop.Value.String()
				if err != nil {
					return nil, err
				}
			case "FLAGS":
				target.Flags, err = prop.Value.String()
				if err != nil {
					return nil, err
				}
				if target.Flags == "AP" {
					target.Status = 1
				}
			case "PRIORITY":
				target.Priority, err = prop.Value.Int()
				if err != nil {
					return nil, err
				}
			case "ATTRS":
				err := parseDestinationAttributes(prop, target)
				if err != nil {
					return nil, err
				}
			case "LATENCY":
				err := parseDestinationLatency(prop, target)
				if err != nil {
					return nil, err
				}
			}
		}

		targets = append(targets, target)
	}
	return targets, nil
}

func parseDestinationLatency(prop binrpc.StructItem, target DispatcherTarget) error {
	latency, err := prop.Value.StructItems()
	if err != nil {
		return err
	}
	for _, attr := range latency {
		switch attr.Key {
		case "AVG":
			target.LatencyAvg, _ = attr.Value.Double()
		case "STD":
			target.LatencyStd, _ = attr.Value.Double()
		case "EST":
			target.LatencyEst, _ = attr.Value.Double()
		case "MAX":
			target.LatencyMax, _ = attr.Value.Double()
		case "TIMEOUT":
			target.LatencyTimeout, _ = attr.Value.Double()
		}
	}
	return nil
}

func parseDestinationAttributes(prop binrpc.StructItem, target DispatcherTarget) error {
	attrs, err := prop.Value.StructItems()
	if err != nil {
		return err
	}
	for _, attr := range attrs {
		switch attr.Key {
		case "BODY":
			target.Body, _ = attr.Value.String()
		case "WEIGHT":
			target.Weight, _ = attr.Value.Int()
		case "RWEIGHT":
			target.RWeight, _ = attr.Value.Int()
		case "SOCKET":
			target.Socket, _ = attr.Value.String()
		}
	}
	return nil
}

func ParseDispatcherMapping(dispatcherMap *[]string, logger log.Logger) map[int]string {
	mapping := make(map[int]string)
	for _, dispatcher := range *dispatcherMap {
		entry := strings.Split(dispatcher, ":")
		if len(entry) != 2 {
			level.Warn(logger).Log("msg", "Invalid dispatcher mapping. Removing the entry", "dispatcher", dispatcher)
			continue
		}
		id, err := strconv.Atoi(entry[0])
		if err != nil {
			level.Warn(logger).Log("msg", "Invalid dispatcher ID. Removing the entry", "dispatcher", dispatcher, "err", err)
			continue
		}
		name := entry[1]
		mapping[id] = name
	}
	return mapping
}
