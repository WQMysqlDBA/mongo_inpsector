// mongodb_exporter
// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package AccessMongoDB

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	exporterPrefix = "mongodb_"
)

type rawMetric struct {
	// Full Qualified Name
	fqName string
	// Help string
	help string
	// Label names
	ln []string
	// Label values
	lv []string
	// Metric value as float64
	val float64
	// Value type
	vt prometheus.ValueType
}
var (
	errCannotHandleType   = fmt.Errorf("don't know how to handle data type")
	errUnexpectedDataType = fmt.Errorf("unexpected data type")
)
//nolint:funlen
func conversions() []conversion {
	return []conversion{
		{
			oldName:          "mongodb_asserts_total",
			newName:          "mongodb_ss_asserts",
			labelConversions: map[string]string{"assert_type": "type"},
		},
		{
			oldName:          "mongodb_connections",
			newName:          "mongodb_ss_connections",
			labelConversions: map[string]string{"conn_type": "state"},
		},
		{
			oldName: "mongodb_connections_metrics_created_total",
			newName: "mongodb_ss_connections_totalCreated",
		},
		{
			oldName: "mongodb_extra_info_page_faults_total",
			newName: "mongodb_ss_extra_info_page_faults",
		},
		{
			oldName: "mongodb_mongod_durability_journaled_megabytes",
			newName: "mongodb_ss_dur_journaledMB",
		},
		{
			oldName: "mongodb_mongod_durability_commits",
			newName: "mongodb_ss_dur_commits",
		},
		{
			oldName: "mongodb_mongod_background_flushing_average_milliseconds",
			newName: "mongodb_ss_backgroundFlushing_average_ms",
		},
		{
			oldName:     "mongodb_mongod_global_lock_client",
			prefix:      "mongodb_ss_globalLock_activeClients",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"readers": "reader",
				"writers": "writer",
				"total":   "total",
			},
		},
		{
			oldName:          "mongodb_mongod_global_lock_current_queue",
			newName:          "mongodb_ss_globalLock_currentQueue",
			labelConversions: map[string]string{"count_type": "type"},
			labelValueConversions: map[string]string{
				"readers": "reader",
				"writers": "writer",
			},
		},
		{
			oldName: "mongodb_instance_local_time",
			newName: "mongodb_start",
		},

		{
			oldName: "mongodb_mongod_instance_uptime_seconds",
			newName: "mongodb_ss_uptime",
		},
		{
			oldName: "mongodb_instance_uptime_seconds",
			newName: "mongodb_ss_uptime",
		},
		{
			oldName: "mongodb_mongod_locks_time_locked_local_microseconds_total",
			newName: "mongodb_ss_locks_Local_acquireCount_[rw]",
		},
		{
			oldName: "mongodb_memory",
			newName: "mongodb_ss_mem_[resident|virtual]",
		},
		{
			oldName:     "mongodb_memory",
			prefix:      "mongodb_ss_mem",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"mapped":            "mapped",
				"mappedWithJournal": "mapped_with_journal",
			},
		},
		{
			oldName:          "mongodb_mongod_metrics_cursor_open",
			newName:          "mongodb_ss_metrics_cursor_open",
			labelConversions: map[string]string{"csr_type": "state"},
		},
		{
			oldName: "mongodb_mongod_metrics_cursor_timed_out_total",
			newName: "mongodb_ss_metrics_cursor_timedOut",
		},
		{
			oldName:          "mongodb_mongod_metrics_document_total",
			newName:          "mongodb_ss_metric_document",
			labelConversions: map[string]string{"doc_op_type": "type"},
		},
		{
			oldName: "mongodb_mongod_metrics_get_last_error_wtime_num_total",
			newName: "mongodb_ss_metrics_getLastError_wtime_num",
		},
		{
			oldName: "mongodb_mongod_metrics_get_last_error_wtimeouts_total",
			newName: "mongodb_ss_metrics_getLastError_wtimeouts",
		},
		{
			oldName:     "mongodb_mongod_metrics_operation_total",
			prefix:      "mongodb_ss_metrics_operation",
			suffixLabel: "state",
			suffixMapping: map[string]string{
				"scanAndOrder":   "scanAndOrder",
				"writeConflicts": "writeConflicts",
			},
		},
		{
			oldName:     "mongodb_mongod_metrics_query_executor_total",
			prefix:      "mongodb_ss_metrics_query",
			suffixLabel: "state",
		},
		{
			oldName: "mongodb_mongod_metrics_record_moves_total",
			newName: "mongodb_ss_metrics_record_moves",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_apply_batches_num_total",
			newName: "mongodb_ss_metrics_repl_apply_batches_num",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_apply_batches_total_milliseconds",
			newName: "mongodb_ss_metrics_repl_apply_batches_totalMillis",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_apply_ops_total",
			newName: "mongodb_ss_metrics_repl_apply_ops",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_buffer_count",
			newName: "mongodb_ss_metrics_repl_buffer_count",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_buffer_max_size_bytes",
			newName: "mongodb_ss_metrics_repl_buffer_maxSizeBytes",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_buffer_size_bytes",
			newName: "mongodb_ss_metrics_repl_buffer_sizeBytes",
		},
		{
			oldName:     "mongodb_mongod_metrics_repl_executor_queue",
			prefix:      "mongodb_ss_metrics_repl_executor_queues",
			suffixLabel: "type",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_executor_unsignaled_events",
			newName: "mongodb_ss_metrics_repl_executor_unsignaledEvents",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_network_bytes_total",
			newName: "mongodb_ss_metrics_repl_network_bytes",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_network_getmores_num_total",
			newName: "mongodb_ss_metrics_repl_network_getmores_num",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_network_getmores_total_milliseconds",
			newName: "mongodb_ss_metrics_repl_network_getmores_totalMillis",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_network_ops_total",
			newName: "mongodb_ss_metrics_repl_network_ops",
		},
		{
			oldName: "mongodb_mongod_metrics_repl_network_readers_created_total",
			newName: "mongodb_ss_metrics_repl_network_readersCreated",
		},
		{
			oldName: "mongodb_mongod_metrics_ttl_deleted_documents_total",
			newName: "mongodb_ss_metrics_ttl_deletedDocuments",
		},
		{
			oldName: "mongodb_mongod_metrics_ttl_passes_total",
			newName: "mongodb_ss_metrics_ttl_passes",
		},
		{
			oldName:     "mongodb_network_bytes_total",
			prefix:      "mongodb_ss_network",
			suffixLabel: "state",
		},
		{
			oldName: "mongodb_network_metrics_num_requests_total",
			newName: "mongodb_ss_network_numRequests",
		},
		{
			oldName:          "mongodb_mongod_op_counters_repl_total",
			newName:          "mongodb_ss_opcountersRepl",
			labelConversions: map[string]string{"legacy_op_type": "type"},
		},
		{
			oldName:          "mongodb_op_counters_total",
			newName:          "mongodb_ss_opcounters",
			labelConversions: map[string]string{"legacy_op_type": "type"},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_blockmanager_blocks_total",
			prefix:      "mongodb_ss_wt_block_manager",
			suffixLabel: "type",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_cache_max_bytes",
			newName: "mongodb_ss_wt_cache_maximum_bytes_configured",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_cache_overhead_percent",
			newName: "mongodb_ss_wt_cache_percentage_overhead",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_concurrent_transactions_available_tickets",
			newName: "mongodb_ss_wt_concurrentTransactions_available",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_concurrent_transactions_out_tickets",
			newName: "mongodb_ss_wt_concurrentTransactions_out",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_concurrent_transactions_total_tickets",
			newName: "mongodb_ss_wt_concurrentTransactions_totalTickets",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_log_records_scanned_total",
			newName: "mongodb_ss_wt_log_records_processed_by_log_scan",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_session_open_cursors_total",
			newName: "mongodb_ss_wt_session_open_cursor_count",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_session_open_sessions_total",
			newName: "mongodb_ss_wt_session_open_session_count",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_transactions_checkpoint_milliseconds_total",
			newName: "mongodb_ss_wt_txn_transaction_checkpoint_total_time_msecs",
		},
		{
			oldName: "mongodb_mongod_wiredtiger_transactions_running_checkpoints",
			newName: "mongodb_ss_wt_txn_transaction_checkpoint_currently_running",
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_transactions_total",
			prefix:      "mongodb_ss_wt_txn_transactions",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"begins":      "begins",
				"checkpoints": "checkpoints",
				"committed":   "committed",
				"rolled_back": "rolled_back",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_blockmanager_bytes_total",
			prefix:      "mongodb_ss_wt_block_manager",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"bytes_read": "read", "mapped_bytes_read": "read_mapped",
				"bytes_written": "written",
			},
		},
		// the 2 metrics bellow have the same prefix.
		{
			oldName:     "mongodb_mongod_wiredtiger_cache_bytes",
			prefix:      "mongodb_ss_wt_cache_bytes",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"currently_in_the_cache":                                 "total",
				"tracked_dirty_bytes_in_the_cache":                       "dirty",
				"tracked_bytes_belonging_to_internal_pages_in_the_cache": " internal_pages",
				"tracked_bytes_belonging_to_leaf_pages_in_the_cache":     "internal_pages",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_cache_bytes_total",
			prefix:      "mongodb_ss_wt_cache",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"bytes_read_into_cache":    "read",
				"bytes_written_from_cache": "written",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_cache_pages",
			prefix:      "mongodb_ss_wt_cache",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"pages_currently_held_in_the_cache": "total",
				"tracked_dirty_pages_in_the_cache":  "dirty",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_cache_pages_total",
			prefix:      "mongodb_ss_wt_cache",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"pages_read_into_cache":    "read",
				"pages_written_from_cache": "written",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_log_records_total",
			prefix:      "mongodb_ss_wt_log",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"log_records_compressed":     "compressed",
				"log_records_not_compressed": "uncompressed",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_log_bytes_total",
			prefix:      "mongodb_ss_wt_log",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"log_bytes_of_payload_data": "payload",
				"log_bytes_written":         "unwritten",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_log_operations_total",
			prefix:      "mongodb_ss_wt_log",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"log_read_operations":                  "read",
				"log_write_operations":                 "write",
				"log_scan_operations":                  "scan",
				"log_scan_records_requiring_two_reads": "scan_double",
				"log_sync_operations":                  "sync",
				"log_sync_dir_operations":              "sync_dir",
				"log_flush_operations":                 "flush",
			},
		},
		{
			oldName:     "mongodb_mongod_wiredtiger_transactions_checkpoint_milliseconds",
			prefix:      "mongodb_ss_wt_txn_transaction_checkpoint",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"min_time_msecs": "min",
				"max_time_msecs": "max",
			},
		},
		{
			oldName:          "mongodb_mongod_global_lock_current_queue",
			prefix:           "mongodb_mongod_global_lock_current_queue",
			labelConversions: map[string]string{"op_type": "type"},
		},
		{
			oldName:          "mongodb_mongod_op_latencies_ops_total",
			newName:          "mongodb_ss_opLatencies_ops",
			labelConversions: map[string]string{"op_type": "type"},
			labelValueConversions: map[string]string{
				"commands": "command",
				"reads":    "read",
				"writes":   "write",
			},
		},
		{
			oldName:          "mongodb_mongod_op_latencies_latency_total",
			newName:          "mongodb_ss_opLatencies_latency",
			labelConversions: map[string]string{"op_type": "type"},
			labelValueConversions: map[string]string{
				"commands": "command",
				"reads":    "read",
				"writes":   "write",
			},
		},
		{
			oldName:          "mongodb_mongod_metrics_document_total",
			newName:          "mongodb_ss_metrics_document",
			labelConversions: map[string]string{"doc_op_type": "state"},
		},
		{
			oldName:     "mongodb_mongod_metrics_query_executor_total",
			prefix:      "mongodb_ss_metrics_queryExecutor",
			suffixLabel: "state",
			suffixMapping: map[string]string{
				"scanned":        "scanned",
				"scannedObjects": "scanned_objects",
			},
		},
		{
			oldName:     "mongodb_memory",
			prefix:      "mongodb_ss_mem",
			suffixLabel: "type",
			suffixMapping: map[string]string{
				"resident": "resident",
				"virtual":  "virtual",
			},
		},
		{
			oldName: "mongodb_mongod_metrics_get_last_error_wtime_total_milliseconds",
			newName: "mongodb_ss_metrics_getLastError_wtime_totalMillis",
		},
		{
			oldName: "mongodb_ss_wt_cache_maximum_bytes_configured",
			newName: "mongodb_mongod_wiredtiger_cache_max_bytes",
		},
		{
			oldName: "mongodb_mongod_db_collections_total",
			newName: "mongodb_dbstats_collections",
		},
		{
			oldName: "mongodb_mongod_db_data_size_bytes",
			newName: "mongodb_dbstats_dataSize",
		},
		{
			oldName: "mongodb_mongod_db_index_size_bytes",
			newName: "mongodb_dbstats_indexSize",
		},
		{
			oldName: "mongodb_mongod_db_indexes_total",
			newName: "mongodb_dbstats_indexes",
		},
		{
			oldName: "mongodb_mongod_db_objects_total",
			newName: "mongodb_dbstats_objects",
		},
	}
}

func appendCompatibleMetric(res []prometheus.Metric, rm *rawMetric) []prometheus.Metric {
	compatibleMetrics := metricRenameAndLabel(rm, conversions())
	if compatibleMetrics == nil {
		return res
	}

	for _, compatibleMetric := range compatibleMetrics {
		metric, err := rawToPrometheusMetric(compatibleMetric)
		if err != nil {
			invalidMetric := prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
			res = append(res, invalidMetric)
			return res
		}

		res = append(res, metric)
	}

	return res
}

//nolint:gochecknoglobals
var (
	// Rules to shrink metric names
	// Please do not change the definitions order: rules are sorted by precedence.
	prefixes = [][]string{
		{"serverStatus.wiredTiger.transaction", "ss_wt_txn"},
		{"serverStatus.wiredTiger", "ss_wt"},
		{"serverStatus", "ss"},
		{"replSetGetStatus", "rs"},
		{"systemMetrics", "sys"},
		{"local.oplog.rs.stats.wiredTiger", "oplog_stats_wt"},
		{"local.oplog.rs.stats", "oplog_stats"},
		{"collstats_storage.wiredTiger", "collstats_storage_wt"},
		{"collstats_storage.indexDetails", "collstats_storage_idx"},
		{"collStats.storageStats", "collstats_storage"},
		{"collStats.latencyStats", "collstats_latency"},
	}

	// This map is used to add labels to some specific metrics.
	// For example, the fields under the serverStatus.opcounters. structure have this
	// signature:
	//
	//    "opcounters": primitive.M{
	//        "insert":  int32(4),
	//        "query":   int32(2118),
	//        "update":  int32(14),
	//        "delete":  int32(22),
	//        "getmore": int32(9141),
	//        "command": int32(67923),
	//    },
	//
	// Applying the renaming rules, serverStatus will become ss but instead of having metrics
	// with the form ss.opcounters.<operation> where operation is each one of the fields inside
	// the structure (insert, query, update, etc), those keys will become labels for the same
	// metric name. The label name is defined as the value for each metric name in the map and
	// the value the label will have is the field name in the structure. Example.
	//
	//   mongodb_ss_opcounters{legacy_op_type="insert"} 4
	//   mongodb_ss_opcounters{legacy_op_type="query"} 2118
	//   mongodb_ss_opcounters{legacy_op_type="update"} 14
	//   mongodb_ss_opcounters{legacy_op_type="delete"} 22
	//   mongodb_ss_opcounters{legacy_op_type="getmore"} 9141
	//   mongodb_ss_opcounters{legacy_op_type="command"} 67923
	//
	nodeToPDMetrics = map[string]string{
		"collStats.storageStats.indexDetails.":            "index_name",
		"globalLock.activeQueue.":                         "count_type",
		"globalLock.locks.":                               "lock_type",
		"serverStatus.asserts.":                           "assert_type",
		"serverStatus.connections.":                       "conn_type",
		"serverStatus.globalLock.currentQueue.":           "count_type",
		"serverStatus.metrics.commands.":                  "cmd_name",
		"serverStatus.metrics.cursor.open.":               "csr_type",
		"serverStatus.metrics.document.":                  "doc_op_type",
		"serverStatus.opLatencies.":                       "op_type",
		"serverStatus.opReadConcernCounters.":             "concern_type",
		"serverStatus.opcounters.":                        "legacy_op_type",
		"serverStatus.opcountersRepl.":                    "legacy_op_type",
		"serverStatus.transactions.commitTypes.":          "commit_type",
		"serverStatus.wiredTiger.concurrentTransactions.": "txn_rw_type",
		"serverStatus.wiredTiger.perf.":                   "perf_bucket",
		"systemMetrics.disks.":                            "device_name",
	}

	// Regular expressions used to make the metric name Prometheus-compatible
	// This variables are global to compile the regexps only once.
	specialCharsRe        = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	repeatedUnderscoresRe = regexp.MustCompile(`__+`)
	dollarRe              = regexp.MustCompile(`\_$`)
)

// prometheusize renames metrics by replacing some prefixes with shorter names
// replace special chars to follow Prometheus metric naming rules and adds the
// exporter name prefix.
func prometheusize(s string) string {
	for _, pair := range prefixes {
		if strings.HasPrefix(s, pair[0]+".") {
			s = pair[1] + strings.TrimPrefix(s, pair[0])
			break
		}
	}

	s = specialCharsRe.ReplaceAllString(s, "_")
	s = dollarRe.ReplaceAllString(s, "")
	s = repeatedUnderscoresRe.ReplaceAllString(s, "_")
	s = strings.TrimPrefix(s, "_")

	return exporterPrefix + s
}

// nameAndLabel checks if there are predefined metric name and label for that metric or
// the standard metrics name should be used in place.
func nameAndLabel(prefix, name string) (string, string) {
	if label, ok := nodeToPDMetrics[prefix]; ok {
		return prometheusize(prefix), label
	}

	return prometheusize(prefix + name), ""
}

// makeRawMetric creates a Prometheus metric based on the parameters we collected by
// traversing the MongoDB structures returned by the collector functions.
func makeRawMetric(prefix, name string, value interface{}, labels map[string]string) (*rawMetric, error) {
	f, err := asFloat64(value)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, nil
	}

	help := metricHelp(prefix, name)

	fqName, label := nameAndLabel(prefix, name)

	rm := &rawMetric{
		fqName: fqName,
		help:   help,
		val:    *f,
		vt:     prometheus.UntypedValue,
		ln:     make([]string, 0, len(labels)),
		lv:     make([]string, 0, len(labels)),
	}

	// Add original labels to the metric
	for k, v := range labels {
		rm.ln = append(rm.ln, k)
		rm.lv = append(rm.lv, v)
	}

	// Add predefined label, if any
	if label != "" {
		rm.ln = append(rm.ln, label)
		rm.lv = append(rm.lv, name)
	}

	return rm, nil
}

func asFloat64(value interface{}) (*float64, error) {
	var f float64
	switch v := value.(type) {
	case bool:
		if v {
			f = 1
		}
	case int:
		f = float64(v)
	case int32:
		f = float64(v)
	case int64:
		f = float64(v)
	case float32:
		f = float64(v)
	case float64:
		f = v
	case primitive.DateTime:
		f = float64(v)
	case primitive.A, primitive.ObjectID, primitive.Timestamp, primitive.Binary, string, []uint8, time.Time:
		return nil, nil
	default:
		return nil, errors.Wrapf(errCannotHandleType, "%T", v)
	}
	return &f, nil
}

func rawToPrometheusMetric(rm *rawMetric) (prometheus.Metric, error) {
	d := prometheus.NewDesc(rm.fqName, rm.help, rm.ln, nil)
	return prometheus.NewConstMetric(d, rm.vt, rm.val, rm.lv...)
}

// metricHelp builds the metric help.
// It is a very very very simple function, but the idea is if the future we want
// to improve the help somehow, there is only one place to change it for the real
// functions and for all the tests.
// Use only prefix or name but not both because 2 metrics cannot have same name but different help.
// For metrics where we labelize some keys, if we put the real metric name here it will be rejected
// by prometheus. For first level metrics, there is no prefix so we should use the metric name or
// the help would be empty.
func metricHelp(prefix, name string) string {
	if prefix != "" {
		return prefix
	}

	return name
}

func makeMetrics(prefix string, m bson.M, labels map[string]string, compatibleMode bool) []prometheus.Metric {
	var res []prometheus.Metric

	if prefix != "" {
		prefix += "."
	}

	for k, val := range m {
		switch v := val.(type) {
		case bson.M:
			res = append(res, makeMetrics(prefix+k, v, labels, compatibleMode)...)
		case map[string]interface{}:
			res = append(res, makeMetrics(prefix+k, v, labels, compatibleMode)...)
		case primitive.A:
			v = []interface{}(v)
			res = append(res, processSlice(prefix, k, v, labels, compatibleMode)...)
		case []interface{}:
			continue
		default:
			rm, err := makeRawMetric(prefix, k, v, labels)
			if err != nil {
				invalidMetric := prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
				res = append(res, invalidMetric)
				continue
			}

			// makeRawMetric returns a nil metric for some data types like strings
			// because we cannot extract data from all types
			if rm == nil {
				continue
			}

			metrics := []*rawMetric{rm}

			if renamedMetrics := metricRenameAndLabel(rm, specialConversions()); renamedMetrics != nil {
				metrics = renamedMetrics
			}

			for _, m := range metrics {
				metric, err := rawToPrometheusMetric(m)
				if err != nil {
					invalidMetric := prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
					res = append(res, invalidMetric)
					continue
				}

				res = append(res, metric)

				if compatibleMode {
					res = appendCompatibleMetric(res, m)
				}
			}
		}
	}

	return res
}

// Extract maps from arrays. Only some structures like replicasets have arrays of members
// and each member is represented by a map[string]interface{}.
func processSlice(prefix, k string, v []interface{}, commonLabels map[string]string, compatibleMode bool) []prometheus.Metric {
	metrics := make([]prometheus.Metric, 0)
	labels := make(map[string]string)
	for name, value := range commonLabels {
		labels[name] = value
	}

	for _, item := range v {
		var s map[string]interface{}

		switch i := item.(type) {
		case map[string]interface{}:
			s = i
		case primitive.M:
			s = map[string]interface{}(i)
		default:
			continue
		}

		// use the replicaset or server name as a label
		if name, ok := s["name"].(string); ok {
			labels["member_idx"] = name
		}
		if state, ok := s["stateStr"].(string); ok {
			labels["member_state"] = state
		}

		metrics = append(metrics, makeMetrics(prefix+k, s, labels, compatibleMode)...)
	}

	return metrics
}

type conversion struct {
	newName               string
	oldName               string
	labelConversions      map[string]string // key: current label, value: old exporter (compatible) label
	labelValueConversions map[string]string // key: current label, value: old exporter (compatible) label
	prefix                string
	suffixLabel           string
	suffixMapping         map[string]string
}
func newToOldMetric(rm *rawMetric, c conversion) *rawMetric {
	oldMetric := &rawMetric{
		fqName: c.oldName,
		help:   rm.help,
		val:    rm.val,
		vt:     rm.vt,
		ln:     make([]string, 0, len(rm.ln)),
		lv:     make([]string, 0, len(rm.lv)),
	}

	for _, val := range rm.lv {
		if newLabelVal, ok := c.labelValueConversions[val]; ok {
			oldMetric.lv = append(oldMetric.lv, newLabelVal)
			continue
		}
		oldMetric.lv = append(oldMetric.lv, val)
	}

	// Some label names should be converted from the new (current) name to the
	// mongodb_exporter v1 compatible name
	for _, newLabelName := range rm.ln {
		// if it should be converted, append the old-compatible name
		if oldLabel, ok := c.labelConversions[newLabelName]; ok {
			oldMetric.ln = append(oldMetric.ln, oldLabel)
			continue
		}
		// otherwise, keep the same label name
		oldMetric.ln = append(oldMetric.ln, newLabelName)
	}

	return oldMetric
}

func metricRenameAndLabel(rm *rawMetric, convs []conversion) []*rawMetric {
	// check if the metric exists in the conversions array.
	// if it exists, it should be converted.
	var result []*rawMetric
	for _, cm := range convs {
		switch {
		case cm.newName != "" && rm.fqName == cm.newName: // first renaming case. See (1)
			result = append(result, newToOldMetric(rm, cm))

		case cm.prefix != "" && strings.HasPrefix(rm.fqName, cm.prefix): // second renaming case. See (2)
			conversionSuffix := strings.TrimPrefix(rm.fqName, cm.prefix)
			conversionSuffix = strings.TrimPrefix(conversionSuffix, "_")

			// Check that also the suffix matches.
			// In the conversion array, there are metrics with the same prefix but the 'old' name varies
			// also depending on the metic suffix
			if _, ok := cm.suffixMapping[conversionSuffix]; ok {
				om := createOldMetricFromNew(rm, cm)
				result = append(result, om)
			}
		}
	}

	return result
}
func createOldMetricFromNew(rm *rawMetric, c conversion) *rawMetric {
	suffix := strings.TrimPrefix(rm.fqName, c.prefix)
	suffix = strings.TrimPrefix(suffix, "_")

	if newSuffix, ok := c.suffixMapping[suffix]; ok {
		suffix = newSuffix
	}

	oldMetric := &rawMetric{
		fqName: c.oldName,
		help:   c.oldName,
		val:    rm.val,
		vt:     rm.vt,
		ln:     []string{c.suffixLabel},
		lv:     []string{suffix},
	}

	return oldMetric
}



// specialConversions returns a list of special conversions we want to implement.
// See: https://jira.percona.com/browse/PMM-6506
func specialConversions() []conversion {
	return []conversion{
		{
			oldName:     "mongodb_ss_opLatencies_ops",
			prefix:      "mongodb_ss_opLatencies",
			suffixLabel: "op_type",
			suffixMapping: map[string]string{
				"commands_ops":     "commands",
				"reads_ops":        "reads",
				"transactions_ops": "transactions",
				"writes_ops":       "writes",
			},
		},
		{
			oldName:     "mongodb_ss_opLatencies_latency",
			prefix:      "mongodb_ss_opLatencies",
			suffixLabel: "op_type",
			suffixMapping: map[string]string{
				"commands_latency":     "commands",
				"reads_latency":        "reads",
				"transactions_latency": "transactions",
				"writes_latency":       "writes",
			},
		},
		// mongodb_ss_wt_concurrentTransactions_read_out
		// mongodb_ss_wt_concurrentTransactions_write_out
		{
			oldName:     "mongodb_ss_wt_concurrentTransactions_out",
			prefix:      "mongodb_ss_wt_concurrentTransactions",
			suffixLabel: "txn_rw",
			suffixMapping: map[string]string{
				"read_out":  "read",
				"write_out": "write",
			},
		},
		// mongodb_ss_wt_concurrentTransactions_read_available
		// mongodb_ss_wt_concurrentTransactions_write_available
		{
			oldName:     "mongodb_ss_wt_concurrentTransactions_available",
			prefix:      "mongodb_ss_wt_concurrentTransactions",
			suffixLabel: "txn_rw",
			suffixMapping: map[string]string{
				"read_available":  "read",
				"write_available": "write",
			},
		},
		// mongodb_ss_wt_concurrentTransactions_read_totalTickets
		// mongodb_ss_wt_concurrentTransactions_write_totalTickets
		{
			oldName:     "mongodb_ss_wt_concurrentTransactions_totalTickets",
			prefix:      "mongodb_ss_wt_concurrentTransactions",
			suffixLabel: "txn_rw",
			suffixMapping: map[string]string{
				"read_totalTickets":  "read",
				"write_totalTickets": "write",
			},
		},
	}
}
