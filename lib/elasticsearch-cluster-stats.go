package mpelasticsearchclusterstats

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("metrics.plugin.elasticsearchclusterstats")

var metricPlace = map[string][]string{
	"docs_count":                  {"indices", "docs", "count"},
	"docs_deleted":                {"indices", "docs", "deleted"},
	"fielddata_size":              {"indices", "fielddata", "memory_size_in_bytes"},
	"query_cache_size":            {"indices", "query_cache", "memory_size_in_bytes"},
	"segments_size":               {"indices", "segments", "memory_in_bytes"},
	"segments_terms_size":         {"indices", "segments", "terms_memory_in_bytes"},
	"segments_stored_fields_size": {"indices", "segments", "stored_fields_memory_in_bytes"},
	"segments_norms_size":         {"indices", "segments", "norms_memory_in_bytes"},
	"segments_points_size":        {"indices", "segments", "points_memory_in_bytes"},
	"segments_doc_values_size":    {"indices", "segments", "doc_values_memory_in_bytes"},
	"segments_index_writer_size":  {"indices", "segments", "index_writer_memory_in_bytes"},
	"segments_version_map_size":   {"indices", "segments", "version_map_memory_in_bytes"},
	"segments_fixed_bit_set_size": {"indices", "segments", "fixed_bit_set_memory_in_bytes"},
	"evictions_fielddata":         {"indices", "fielddata", "evictions"},
	"evictions_query_cache":       {"indices", "query_cache", "evictions"},
}

func getFloatValue(s map[string]interface{}, keys []string) (float64, error) {
	var val float64
	sm := s
	for i, k := range keys {
		if i+1 < len(keys) {
			switch sm[k].(type) {
			case map[string]interface{}:
				sm = sm[k].(map[string]interface{})
			default:
				return 0, errors.New("Cannot handle as a hash")
			}
		} else {
			switch sm[k].(type) {
			case float64:
				val = sm[k].(float64)
			default:
				return 0, errors.New("Not float64")
			}
		}
	}

	return val, nil
}

// ElasticsearchPlugin mackerel plugin
type ElasticsearchPlugin struct {
	URI    string
	Prefix string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p *ElasticsearchPlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		p.Prefix = "elasticsearchclusterstats"
	}
	return p.Prefix
}

// GraphDefinition interface for mackerelplugin
func (p *ElasticsearchPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(p.Prefix)
	return map[string]mp.Graphs{
		"indices.docs": {
			Label: labelPrefix + " Indices Docs",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "docs_count", Label: "Count", Stacked: true},
				{Name: "docs_deleted", Label: "Deleted", Stacked: true},
			},
		},
		"indices.memory_size": {
			Label: labelPrefix + " Indices Memory Size",
			Unit:  "bytes",
			Metrics: []mp.Metrics{
				{Name: "fielddata_size", Label: "Fielddata", Stacked: true},
				{Name: "query_cache_size", Label: "Query Cache", Stacked: true},
				{Name: "segments_size", Label: "Lucene Segments", Stacked: true},
				{Name: "segments_terms_size", Label: "Lucene Segments Term", Stacked: true},
				{Name: "segments_stored_fields_size", Label: "Lucene Segments Stored Fields", Stacked: true},
				{Name: "segments_norms_size", Label: "Lucene Segments Norms", Stacked: true},
				{Name: "segments_points_size", Label: "Lucene Segments Points", Stacked: true},
				{Name: "segments_doc_values_size", Label: "Lucene Segments Doc Values", Stacked: true},
				{Name: "segments_index_writer_size", Label: "Lucene Segments Index Writer", Stacked: true},
				{Name: "segments_version_map_size", Label: "Lucene Segments Version Map", Stacked: true},
				{Name: "segments_fixed_bit_set_size", Label: "Lucene Segments Fixed Bit Set", Stacked: true},
			},
		},
		"indices.evictions": {
			Label: labelPrefix + " Indices Evictions",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "evictions_fielddata", Label: "Fielddata", Diff: true},
				{Name: "evictions_filter_cache", Label: "Query Cache", Diff: true},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p *ElasticsearchPlugin) FetchMetrics() (map[string]float64, error) {
	req, err := http.NewRequest(http.MethodGet, p.URI+"/_cluster/stats", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mackerel-plugin-elasticsearch")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics := make(map[string]float64)
	decoder := json.NewDecoder(resp.Body)

	var s map[string]interface{}
	err = decoder.Decode(&s)
	if err != nil {
		return nil, err
	}

	for k, v := range metricPlace {
		val, err := getFloatValue(s, v)
		if err != nil {
			logger.Errorf("Failed to find '%s': %s", k, err)
			continue
		}
		metrics[k] = val
	}

	return metrics, nil
}

// Do the plugin
func Do() {
	optScheme := flag.String("scheme", "http", "Scheme")
	optHost := flag.String("host", "localhost", "Host")
	optPort := flag.String("port", "9200", "Port")
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	plugin := mp.NewMackerelPlugin(&ElasticsearchPlugin{
		URI:    fmt.Sprintf("%s://%s:%s", *optScheme, *optHost, *optPort),
		Prefix: *optPrefix,
	})
	if *optTempfile != "" {
		plugin.Tempfile = *optTempfile
	} else {
		plugin.SetTempfileByBasename(fmt.Sprintf("mackerel-plugin-elasticsearch-cluster-stats-%s-%s", *optHost, *optPort))
	}
	plugin.Run()
}
