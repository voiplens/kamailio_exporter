# Prometheus Exporter for Kamailio

[![Project Banner](.github/banner.png?raw=true)](https://angarium.io)

[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](https://github.com/angarium-cloud/kamailio_exporter/blob/master/LICENSE.txt)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/angarium-cloud/kamailio_exporter/build.yml?style=for-the-badge)
![GitHub Release Date - Published_At](https://img.shields.io/github/release-date/angarium-cloud/kamailio_exporter?style=for-the-badge)

A Prometheus exporter for [Kamailio](https://www.kamailio.org/) which exports the following metrics:

- Core metrics (Shared memory, Request/Reply, TCP, Dialog SL, TMX)
- Uptime and core information about Kamailio
- Status of each process running
- Extra TCP metrics
- Dispatcher list status
- Dialog metrics
- Number of dialogs belonging to a profile.
- Htable status and metrics
- Extra Private memory metrics
- RTPengine status
- Additional SL module Stats
- Additional TM module Stats
- TLS metrics

This project started as a fork of the [pascomnet/kamailio_exporter](https://github.com/pascomnet/kamailio_exporter).

## Kamailio configuration

The Exporter needs a [BINRPC](http://kamailio.org/docs/modules/stable/modules/ctl.html) connection to Kamailio.
You have to load and configure the [CTL](http://kamailio.org/docs/modules/stable/modules/ctl.html) module.

If you run the Exporter and Kamailio on the same Machine, it's recommended to use a Unix socket for the connection.
The path for the socket defaults to "unix:/var/run/kamailio/kamailio_ctl" and can be used out of the box.

Depending on your deployment, you might want to open a TCP socket on a _private or firewalled_ interface.
This allows you, for example, to run the Exporter as a Sidecar to your Kamailio Container in a Dockerized environment.

```
modparam("ctl", "binrpc", "tcp:192.168.1.10:2046")
```

## Running the Exporter

Download or build the kamailio_exporter binary and start it. If you do so, it'll try to reach Kamailio on the default unix domain socket /var/run/kamailio/kamailio_ctl. The exporter runs on port `9494` all available interfaces and exports all the metrics on the `/metrics` path.

### Configuration

You can configure the exporter using the following flags:

- `--kamailio.binrpc-uri="`: BINRPC URI on which to scrape kamailio. Defaults to `unix:///var/run/kamailio/kamailio_ctl"` for TCP use `"tcp://192.168.1.10:2046"` format.
- `--kamailio.timeout`: Timeout for trying to get stats from Kamailio using BINRPC. Default to `5s`.
- `--kamailio.custom-metrics-url`: URL to request user-defined metrics from Kamailio.
- `--collector.dispatcher.mapping`: Map a Dispatcher ID to a Name using the "ID:NAME" format. E.g. "100:Genesys".
- `--collector.dialog.profiles`: Select dialog profiles to query.
- `--web.telemetry-path`: Path under which to expose metrics. Defaults to `/metrics`.
- `--web.rtp-telemetry-path`: Path under which to expose rtpengine metrics.
- `--[no-]web.systemd-socket`: Use systemd socket activation listeners instead of port listeners (Linux only).
- `--web.listen-address"`: Addresses on which to expose metrics and web interface. Repeatable for multiple addresses. Defaults to `:9494`.
- `--web.config.file`: Path to a configuration file that can enable TLS or authentication. See: https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
- `--log.level`: Only log messages with the given severity or above. One of: [`debug`, `info`, `warn`, `error`]. Defaults to `info`.
- `--log.format`: Output format of log messages. One of: [`logfmt`, `json`]. Defaults to `logfmt`.
- `--[no-]version`: Show application version.

Test that the exporter is running and can collect metrics from Kamailio using the following command: `curl -s http://localhost:9494/metrics | grep kamailio_up`.
The output should return this:

```
# HELP kamailio_up kamailio_exporter: Whether the Kamailio endpoint is up.
# TYPE kamailio_up gauge
kamailio_up 1
```

If the value of the `kamailio_up` metrics is `1`, the exporter can connect to Kamailio, and it collects further metrics.

## Exported metrics

### Default stats metrics

These metrics are generated from the `stats.fetch all` command.

```
# HELP kamailio_bad_msg_hdr Messages with bad message header
# TYPE kamailio_bad_msg_hdr counter
kamailio_bad_msg_hdr 0
# HELP kamailio_bad_uri_total Messages with bad uri
# TYPE kamailio_bad_uri_total counter
kamailio_bad_uri_total 0
# HELP kamailio_core_rcv_reply_total Received replies by code
# TYPE kamailio_core_rcv_reply_total counter
kamailio_core_rcv_reply_total{code="18x"} 0
kamailio_core_rcv_reply_total{code="1xx"} 0
kamailio_core_rcv_reply_total{code="2xx"} 0
kamailio_core_rcv_reply_total{code="3xx"} 0
kamailio_core_rcv_reply_total{code="401"} 0
kamailio_core_rcv_reply_total{code="404"} 0
kamailio_core_rcv_reply_total{code="407"} 0
kamailio_core_rcv_reply_total{code="480"} 0
kamailio_core_rcv_reply_total{code="486"} 0
kamailio_core_rcv_reply_total{code="4xx"} 0
kamailio_core_rcv_reply_total{code="5xx"} 0
kamailio_core_rcv_reply_total{code="6xx"} 0
# HELP kamailio_core_rcv_request_total Received requests by method
# TYPE kamailio_core_rcv_request_total counter
kamailio_core_rcv_request_total{method="ack"} 0
kamailio_core_rcv_request_total{method="bye"} 0
kamailio_core_rcv_request_total{method="cancel"} 0
kamailio_core_rcv_request_total{method="info"} 0
kamailio_core_rcv_request_total{method="invite"} 0
kamailio_core_rcv_request_total{method="message"} 0
kamailio_core_rcv_request_total{method="notify"} 0
kamailio_core_rcv_request_total{method="options"} 0
kamailio_core_rcv_request_total{method="prack"} 0
kamailio_core_rcv_request_total{method="publish"} 0
kamailio_core_rcv_request_total{method="refer"} 0
kamailio_core_rcv_request_total{method="register"} 0
kamailio_core_rcv_request_total{method="subscribe"} 0
kamailio_core_rcv_request_total{method="unsupported"} 0
kamailio_core_rcv_request_total{method="update"} 0
# HELP kamailio_core_reply_total Reply counters
# TYPE kamailio_core_reply_total counter
kamailio_core_reply_total{type="drop"} 0
kamailio_core_reply_total{type="err"} 0
kamailio_core_reply_total{type="fwd"} 0
kamailio_core_reply_total{type="rcv"} 0
# HELP kamailio_core_request_total Request counters
# TYPE kamailio_core_request_total counter
kamailio_core_request_total{method="drop"} 0
kamailio_core_request_total{method="err"} 0
kamailio_core_request_total{method="fwd"} 0
kamailio_core_request_total{method="rcv"} 0
# HELP kamailio_dns_failed_request_total Failed dns requests
# TYPE kamailio_dns_failed_request_total counter
kamailio_dns_failed_request_total 0
# HELP kamailio_shm_bytes Shared memory sizes
# TYPE kamailio_shm_bytes gauge
kamailio_shm_bytes{type="free"} 6.3184376e+07
kamailio_shm_bytes{type="max_used"} 3.924488e+06
kamailio_shm_bytes{type="real_used"} 3.924488e+06
kamailio_shm_bytes{type="total"} 6.7108864e+07
kamailio_shm_bytes{type="used"} 3.030808e+06
# HELP kamailio_shm_fragments Shared memory fragment count
# TYPE kamailio_shm_fragments gauge
kamailio_shm_fragments 1
# HELP kamailio_sl_reply_total Stateless replies by code
# TYPE kamailio_sl_reply_total counter
kamailio_sl_reply_total{code="1xx"} 0
kamailio_sl_reply_total{code="200"} 0
kamailio_sl_reply_total{code="202"} 0
kamailio_sl_reply_total{code="2xx"} 0
kamailio_sl_reply_total{code="300"} 0
kamailio_sl_reply_total{code="301"} 0
kamailio_sl_reply_total{code="302"} 0
kamailio_sl_reply_total{code="3xx"} 0
kamailio_sl_reply_total{code="400"} 0
kamailio_sl_reply_total{code="401"} 0
kamailio_sl_reply_total{code="403"} 0
kamailio_sl_reply_total{code="404"} 0
kamailio_sl_reply_total{code="407"} 0
kamailio_sl_reply_total{code="408"} 0
kamailio_sl_reply_total{code="483"} 0
kamailio_sl_reply_total{code="4xx"} 0
kamailio_sl_reply_total{code="500"} 0
kamailio_sl_reply_total{code="5xx"} 0
kamailio_sl_reply_total{code="6xx"} 0
# HELP kamailio_sl_type_total Stateless replies by type
# TYPE kamailio_sl_type_total counter
kamailio_sl_type_total{type="failure"} 0
kamailio_sl_type_total{type="received_ack"} 0
kamailio_sl_type_total{type="sent_err_reply"} 0
kamailio_sl_type_total{type="sent_reply"} 0
kamailio_sl_type_total{type="xxx_reply"} 0
# HELP kamailio_tcp_connections Opened TCP connections
# TYPE kamailio_tcp_connections gauge
kamailio_tcp_connections 0
# HELP kamailio_tcp_total TCP connection counters
# TYPE kamailio_tcp_total counter
kamailio_tcp_total{type="con_reset"} 0
kamailio_tcp_total{type="con_timeout"} 0
kamailio_tcp_total{type="connect_failed"} 0
kamailio_tcp_total{type="connect_success"} 0
kamailio_tcp_total{type="established"} 0
kamailio_tcp_total{type="local_reject"} 0
kamailio_tcp_total{type="passive_open"} 0
kamailio_tcp_total{type="send_timeout"} 0
kamailio_tcp_total{type="sendq_full"} 0
# HELP kamailio_tcp_writequeue TCP write queue size
# TYPE kamailio_tcp_writequeue gauge
kamailio_tcp_writequeue 0
```

### Pkg / Private memory metrics

These metrics are generated from the `pkg.stats` command.
A series of metrics is exported for each Kamailio child process:

```
# HELP kamailio_pkgmem_frags Private memory total frags
# TYPE kamailio_pkgmem_frags gauge
kamailio_pkgmem_frags{entry="0",pid="1"} 254
kamailio_pkgmem_frags{entry="1",pid="7"} 248
# HELP kamailio_pkgmem_free Private memory free
# TYPE kamailio_pkgmem_free gauge
kamailio_pkgmem_free{entry="0",pid="1"} 1.1730984e+07
kamailio_pkgmem_free{entry="1",pid="7"} 1.1727304e+07
# HELP kamailio_pkgmem_real Private memory real used
# TYPE kamailio_pkgmem_real gauge
kamailio_pkgmem_real{entry="0",pid="1"} 5.046232e+06
kamailio_pkgmem_real{entry="1",pid="7"} 5.049912e+06
# HELP kamailio_pkgmem_size Private memory total size
# TYPE kamailio_pkgmem_size gauge
kamailio_pkgmem_size{entry="0",pid="1"} 1.6777216e+07
kamailio_pkgmem_size{entry="1",pid="7"} 1.6777216e+07
# HELP kamailio_pkgmem_used Private memory used
# TYPE kamailio_pkgmem_used gauge
kamailio_pkgmem_used{entry="0",pid="1"} 3.829424e+06
kamailio_pkgmem_used{entry="1",pid="7"} 3.830712e+06
```

### Core Processes status

These metrics are generated from the `core.psa` command.

```
# HELP kamailio_core_process_status Status of each process running in Kamailio
# TYPE kamailio_core_process_status gauge
kamailio_core_process_status{description="main process - attendant",index="0",pid="1",rank="0"} 1
kamailio_core_process_status{description="udp receiver child=0 sock=172.16.105.10:5060 (172.16.104.10:5060)",index="1",pid="7",rank="1"} 1
```

### Core runtime info

These metrics are generated from the `core.runinfo` command.

```
# HELP kamailio_core_uptime Uptime in seconds
# TYPE kamailio_core_uptime gauge
kamailio_core_uptime{compiled="22:28:09 Nov  8 2023",compiler="gcc 13.2.1",version="5.7.2"} 2352
```

### Core TCP/TLS stats

These metrics are generated from the `core.tcp_info` command.

```
# HELP kamailio_tcp_max_connections TCP connection limit
# TYPE kamailio_tcp_max_connections gauge
kamailio_tcp_max_connections 16384
# HELP kamailio_tcp_readers TCP readers
# TYPE kamailio_tcp_readers gauge
kamailio_tcp_readers 8
# HELP kamailio_tls_connections Opened TLS connections
# TYPE kamailio_tls_connections gauge
kamailio_tls_connections 0
# HELP kamailio_tls_max_connections TLS connection limit
# TYPE kamailio_tls_max_connections gauge
kamailio_tls_max_connections 16384
```

### Dispatcher List stats

These metrics are generated from the `dispatcher.list` command.
Use the `--collector.dispatcher.mapping` flag to map a dispatcher Set ID to a Name using the `"ID:NAME"` format. You will need to repeat the option for each mapping. As an example: `kamailio_exporter --collector.dispatcher.mapping="200:Carrier 1" --collector.dispatcher.mapping="400:Carrier 2"`.
Without this option the `set_name` label will always be set to blank.

```
# HELP kamailio_dispatcher_list_target Target status.
# TYPE kamailio_dispatcher_list_target gauge
kamailio_dispatcher_list_target{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 1
kamailio_dispatcher_list_target{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 1
# HELP kamailio_dispatcher_list_target_flags_status Target flags.
# TYPE kamailio_dispatcher_list_target_flags_status gauge
kamailio_dispatcher_list_target_flags_status{destination="sip:172.16.105.138:5070;transport=tcp",flags="AP",set_id="200",set_name="Carrier 1"} 1
kamailio_dispatcher_list_target_flags_status{destination="sip:172.16.106.128:5060",flags="AP",set_id="400",set_name="Carrier 2"} 1
# HELP kamailio_dispatcher_list_target_latency_avg Target Latency Average.
# TYPE kamailio_dispatcher_list_target_latency_avg gauge
kamailio_dispatcher_list_target_latency_avg{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_latency_avg{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_latency_est Target Latency.
# TYPE kamailio_dispatcher_list_target_latency_est gauge
kamailio_dispatcher_list_target_latency_est{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_latency_est{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_latency_max Target Latency.
# TYPE kamailio_dispatcher_list_target_latency_max gauge
kamailio_dispatcher_list_target_latency_max{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_latency_max{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_latency_std Target Latency.
# TYPE kamailio_dispatcher_list_target_latency_std gauge
kamailio_dispatcher_list_target_latency_std{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_latency_std{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_latency_timeout Target Latency.
# TYPE kamailio_dispatcher_list_target_latency_timeout gauge
kamailio_dispatcher_list_target_latency_timeout{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_latency_timeout{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_priority Target Priority.
# TYPE kamailio_dispatcher_list_target_priority gauge
kamailio_dispatcher_list_target_priority{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_priority{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 50
# HELP kamailio_dispatcher_list_target_rweight Target rweight.
# TYPE kamailio_dispatcher_list_target_rweight gauge
kamailio_dispatcher_list_target_rweight{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_rweight{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
# HELP kamailio_dispatcher_list_target_weight Target Weight.
# TYPE kamailio_dispatcher_list_target_weight gauge
kamailio_dispatcher_list_target_weight{destination="sip:172.16.105.138:5070;transport=tcp",set_id="200",set_name="Carrier 1"} 0
kamailio_dispatcher_list_target_weight{destination="sip:172.16.106.128:5060",set_id="400",set_name="Carrier 2"} 0
```

### Dialog stats

These metrics are generated from the `dlg.stats_active` command.

```
# HELP kamailio_dlg_stats_active_all Dialog all.
# TYPE kamailio_dlg_stats_active_all gauge
kamailio_dlg_stats_active_all 0
# HELP kamailio_dlg_stats_active_answering Dialog answering.
# TYPE kamailio_dlg_stats_active_answering gauge
kamailio_dlg_stats_active_answering 0
# HELP kamailio_dlg_stats_active_connecting Dialog connecting.
# TYPE kamailio_dlg_stats_active_connecting gauge
kamailio_dlg_stats_active_connecting 0
# HELP kamailio_dlg_stats_active_ongoing Dialog ongoing.
# TYPE kamailio_dlg_stats_active_ongoing gauge
kamailio_dlg_stats_active_ongoing 0
# HELP kamailio_dlg_stats_active_starting Dialog starting.
# TYPE kamailio_dlg_stats_active_starting gauge
kamailio_dlg_stats_active_starting 0
```

Use the `--collector.dialog.profiles` flag to collect the size of a dialog profile. For example: `kamailio_exporter --collector.dialog.profiles="PROVIDER_A_IN" --collector.dialog.profiles="PROVIDER_A_OUT"`.

```
# HELP kamailio_dlg_profile_get_size_dialog Current number of dialogs belonging to a profile.
# TYPE kamailio_dlg_profile_get_size_dialog gauge
kamailio_dlg_profile_get_size_dialog{profile="PROVIDER_A_IN"} 0
kamailio_dlg_profile_get_size_dialog{profile="PROVIDER_A_OUT"} 0
```

### HTables stats

These metrics are generated from the `htable.listTables` and `htable.stats` commands.

```
# HELP kamailio_htable_auto_expire_seconds Time in seconds to delete an item from a hash table if no update was done to it
# TYPE kamailio_htable_auto_expire_seconds gauge
kamailio_htable_auto_expire_seconds{name="trunkcontrol"} 43200
kamailio_htable_auto_expire_seconds{name="threevpn"} 43200
# HELP kamailio_htable_db_mode_status Htable write back to db table
# TYPE kamailio_htable_db_mode_status gauge
kamailio_htable_db_mode_status{dbtable="",name="trunkcontrol"} 0
kamailio_htable_db_mode_status{dbtable="",name="threevpn"} 0
# HELP kamailio_htable_dmq_replicate_status DMQ Replicate status
# TYPE kamailio_htable_dmq_replicate_status gauge
kamailio_htable_dmq_replicate_status{name="trunkcontrol"} 0
kamailio_htable_dmq_replicate_status{name="threevpn"} 0
# HELP kamailio_htable_items_per_slots_max Max number of items per slot in the htable
# TYPE kamailio_htable_items_per_slots_max gauge
kamailio_htable_items_per_slots_max{name="trunkcontrol"} 0
kamailio_htable_items_per_slots_max{name="threevpn"} 0
# HELP kamailio_htable_items_per_slots_min Min number of items per slot in the htable
# TYPE kamailio_htable_items_per_slots_min gauge
kamailio_htable_items_per_slots_min{name="trunkcontrol"} 0
kamailio_htable_items_per_slots_min{name="threevpn"} 0
# HELP kamailio_htable_items_total Total number of items stored in the htable
# TYPE kamailio_htable_items_total gauge
kamailio_htable_items_total{name="trunkcontrol"} 0
kamailio_htable_items_total{name="threevpn"} 0
# HELP kamailio_htable_slots_total Number of slots in the htable
# TYPE kamailio_htable_slots_total gauge
kamailio_htable_slots_total{name="trunkcontrol"} 256
kamailio_htable_slots_total{name="threevpn"} 256
# HELP kamailio_htable_update_expire_status Update Expire status
# TYPE kamailio_htable_update_expire_status gauge
kamailio_htable_update_expire_status{name="trunkcontrol"} 1
kamailio_htable_update_expire_status{name="threevpn"} 1
```

### RTPEngine connection status

These metrics are generated from the `rtpengine.show` command.

```
# HELP kamailio_rtpengine_enabled rtpengine connection status
# TYPE kamailio_rtpengine_enabled gauge
kamailio_rtpengine_enabled{index="0",set="0",url="udp://172.16.105.20:22223",weight="1"} 1
```

### Stateless UA Server stats

These metrics are generated from the `sl.stats` command.

```
# HELP kamailio_sl_stats_codes_total Per-code counters.
# TYPE kamailio_sl_stats_codes_total counter
kamailio_sl_stats_codes_total{code="200"} 0
kamailio_sl_stats_codes_total{code="202"} 0
kamailio_sl_stats_codes_total{code="2xx"} 0
kamailio_sl_stats_codes_total{code="300"} 0
kamailio_sl_stats_codes_total{code="301"} 0
kamailio_sl_stats_codes_total{code="302"} 0
kamailio_sl_stats_codes_total{code="3xx"} 0
kamailio_sl_stats_codes_total{code="400"} 0
kamailio_sl_stats_codes_total{code="401"} 0
kamailio_sl_stats_codes_total{code="403"} 0
kamailio_sl_stats_codes_total{code="404"} 0
kamailio_sl_stats_codes_total{code="407"} 0
kamailio_sl_stats_codes_total{code="408"} 0
kamailio_sl_stats_codes_total{code="483"} 0
kamailio_sl_stats_codes_total{code="4xx"} 0
kamailio_sl_stats_codes_total{code="500"} 0
kamailio_sl_stats_codes_total{code="5xx"} 0
kamailio_sl_stats_codes_total{code="6xx"} 0
kamailio_sl_stats_codes_total{code="xxx"} 0
```

### SIP Transaction stats

These metrics are generated from the `tm.stats` command.

```
# HELP kamailio_tm_stats_codes_total Per-code counters.
# TYPE kamailio_tm_stats_codes_total counter
kamailio_tm_stats_codes_total{code="2xx"} 2214
kamailio_tm_stats_codes_total{code="3xx"} 0
kamailio_tm_stats_codes_total{code="4xx"} 0
kamailio_tm_stats_codes_total{code="5xx"} 0
kamailio_tm_stats_codes_total{code="6xx"} 0
# HELP kamailio_tm_stats_created_total Created transactions.
# TYPE kamailio_tm_stats_created_total counter
kamailio_tm_stats_created_total 2214
# HELP kamailio_tm_stats_current Current transactions.
# TYPE kamailio_tm_stats_current gauge
kamailio_tm_stats_current 3
# HELP kamailio_tm_stats_delayed_free_total Delayed free transactions.
# TYPE kamailio_tm_stats_delayed_free_total counter
kamailio_tm_stats_delayed_free_total 0
# HELP kamailio_tm_stats_freed_total Freed transactions.
# TYPE kamailio_tm_stats_freed_total counter
kamailio_tm_stats_freed_total 2211
# HELP kamailio_tm_stats_local_total Total local transactions.
# TYPE kamailio_tm_stats_local_total counter
kamailio_tm_stats_local_total 2214
# HELP kamailio_tm_stats_rpl_generated_total Number of reply generated.
# TYPE kamailio_tm_stats_rpl_generated_total counter
kamailio_tm_stats_rpl_generated_total 0
# HELP kamailio_tm_stats_rpl_received_total Number of reply received.
# TYPE kamailio_tm_stats_rpl_received_total counter
kamailio_tm_stats_rpl_received_total 2214
# HELP kamailio_tm_stats_rpl_sent_total Number of reply sent.
# TYPE kamailio_tm_stats_rpl_sent_total counter
kamailio_tm_stats_rpl_sent_total 2214
# HELP kamailio_tm_stats_total Total transactions.
# TYPE kamailio_tm_stats_total counter
kamailio_tm_stats_total 2214
# HELP kamailio_tm_stats_waiting Waiting transactions.
# TYPE kamailio_tm_stats_waiting gauge
kamailio_tm_stats_waiting 3
```

## Kamailio xhttp_prom metrics

If your kamailio server supports it and is configured correctly, the exporter can query Kamailio's xhttp_prom metrics and combine them with the other metrics generated by this exporter.
See `--customKamailioMetricsURL`.

This allows you to configure and fill custom counters and gauges using the built-in Prometheus functions in Kamailio, with labels.

If you want to use it, you need to enable and configure the `xhttp` and `xhttp_prom` modules in Kamailio. See [Kamailio module docs](https://kamailio.org/docs/modules/devel/modules/xhttp_prom.html) as a starting point. Be careful when setting the `xhttp_prom_stats` parameter. You might get unwanted results if you expose too many native metrics.

## Scripted metrics

Often you might want to record some values from your own business logic. As usual in the Kamailio ecosystem,
there is already a module for this purpose: [statistics](http://kamailio.org/docs/modules/stable/modules/statistics.html)

Statistics Module can be used both from Kamailio native scripts and all KEMI Languages, e.g. Lua or Python.

Configuration and usage is quite simple. Of course you need to load it:

```
loadmodule "statistics.so"
```

Next, all to-be-exported metrics have to be declared as statistic variable:

```
modparam("statistics", "variable", "my_custom_value_total")
```

Finally, in some route block, you have to populate the statistic variable with a value:

```
update_stat("my_custom_value_total", "+1");
```

A scraped metric will look like this:

```
# HELP kamailio_my_custom_value_totalScripted metric my_custom_value_total
# TYPE kamailio_my_custom_value_total counter
kamailio_my_custom_value_total 1
```

### Scripted metric details

- the statistic variable name is prefixed by "kamailio\_" and changed to lower-case
- a suffix of "\_total", "\_seconds" or "\_bytes" will export a Prometheus Counter, omitting the suffix produces a Prometheus Gauge, see [metric types](https://prometheus.io/docs/concepts/metric_types/).

## Building from source

To build the Kamailio Exporter from source code, you need a working Go development environmemt with a minimum go version 1.21.
Start by cloning the repository:

```sh
git clone https://github.com/angarium-cloud/kamailio_exporter.git
cd kamailio_exporter
```

Then, simply run make to build the executable:

```sh
make
```

This exporter uses the common prometheus tooling to build and run some tests.

## Acknowledgements

Kudos to Florent Chauveau for his Golang BINRPC implementation: https://github.com/florentchauveau/go-kamailio-binrpc.
Also noteworthy is that he provides an alternative implementation for scraping Kamailio statistics: https://github.com/florentchauveau/kamailio_exporter
