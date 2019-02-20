# mackerel-plugin-elasticsearch-cluster-stats

Elasticsearch Cluster Stats custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-elasticsearch-cluster-stats [-scheme=<'http'|'https'>] [-host=<host>] [-port=<port>] [-metric-key-prefix=<prefix>] [-tempfile=<tempfile>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.elasticsearch-cluster-stats]
command = "/path/to/mackerel-plugin-elasticsearch-cluster-stats"
```
