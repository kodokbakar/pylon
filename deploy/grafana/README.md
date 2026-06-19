# Pylon Grafana Dashboards

This directory contains Grafana provisioning configuration and dashboards for Pylon observability.

## Provisioning Structure

```text
deploy/grafana/
├── provisioning/
│   ├── datasources/
│   │   └── datasource.yml
│   └── dashboards/
│       └── dashboard.yml
└── dashboards/
    ├── pylon-overview.json
    ├── api-gateway.json
    ├── microservices.json
    ├── kafka.json
    ├── infrastructure.json
    └── business-metrics.json
````

## Dashboards

| Dashboard              | Purpose                                                                               |
| ---------------------- | ------------------------------------------------------------------------------------- |
| Pylon Overview         | High-level HTTP, gRPC, WebSocket, users, and messages overview                        |
| Pylon API Gateway      | HTTP request rate, latency, in-flight requests, WebSocket activity, and rate limiting |
| Pylon Microservices    | gRPC request rate, latency, errors, and Kubernetes pod resource usage                 |
| Pylon Kafka            | Kafka publish/consume metrics, publish latency, consumer lag, and broker metrics      |
| Pylon Infrastructure   | PostgreSQL, Redis, and Kubernetes infrastructure metrics                              |
| Pylon Business Metrics | Users online, rooms created, messages sent, and active WebSocket connections          |

## Data Source

The dashboards expect a Grafana datasource with UID `Prometheus`.

Provisioned datasource:

```yaml
name: Prometheus
uid: Prometheus
type: prometheus
url: http://prometheus:9090
```

## Required Metrics

Pylon application dashboards use metrics exported directly by the services:

| Metric                                 | Source                               |
| -------------------------------------- | ------------------------------------ |
| `pylon_http_requests_total`            | API Gateway HTTP middleware          |
| `pylon_http_request_duration_seconds`  | API Gateway HTTP middleware          |
| `pylon_http_requests_in_flight`        | API Gateway HTTP middleware          |
| `pylon_grpc_requests_total`            | Service gRPC middleware              |
| `pylon_grpc_request_duration_seconds`  | Service gRPC middleware              |
| `pylon_kafka_messages_published_total` | Chat Service Kafka producer          |
| `pylon_kafka_messages_consumed_total`  | Notification Service Kafka consumer  |
| `pylon_kafka_publish_duration_seconds` | Chat Service Kafka producer          |
| `pylon_websocket_connections_active`   | API Gateway WebSocket handler        |
| `pylon_messages_sent_total`            | Chat Service message flow            |
| `pylon_rooms_created_total`            | Room Service room creation flow      |
| `pylon_users_online`                   | Presence Service online/offline flow |

## External Exporter Requirements

Some panels depend on infrastructure metrics that are not exported directly by Pylon. These panels are valid, but they will show `No data` until the matching exporters are installed and scraped by Prometheus.

| Dashboard      | Metrics                                                                                                          | Required Exporter                    |
| -------------- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| Infrastructure | `pg_stat_activity_count`, `pg_stat_database_*`                                                                   | postgres_exporter                    |
| Infrastructure | `redis_memory_used_bytes`, `redis_keyspace_hits_total`, `redis_keyspace_misses_total`, `redis_connected_clients` | redis_exporter                       |
| Infrastructure | `kube_pod_*`, `kube_horizontalpodautoscaler_*`                                                                   | kube-state-metrics                   |
| Microservices  | `container_cpu_usage_seconds_total`, `container_memory_working_set_bytes`, `container_network_*`                 | cAdvisor / kubelet metrics           |
| Kafka          | `kafka_consumergroup_lag`                                                                                        | kafka-exporter                       |
| Kafka          | `kafka_server_brokertopicmetrics_*`                                                                              | Kafka JMX exporter or kafka-exporter |

## Kafka Dashboard Variables

The Kafka dashboard includes a `consumer_group` template variable.

Default behavior:

* `All` shows all scraped consumer groups.
* Specific consumer groups can be selected from Grafana.
* The consumer lag panel uses `group=~"$consumer_group"` instead of a hardcoded consumer group name.

## Refresh Interval

All dashboards use:

```text
refresh: 30s
```

## Alerting Notes

Alerting rules are intentionally not provisioned in this dashboard-only change. Production alerting should be added in a dedicated follow-up issue after the Prometheus and Alertmanager deployment model is finalized.

Recommended alerts:

| Alert                | Suggested Condition                                |
| -------------------- | -------------------------------------------------- |
| High HTTP error rate | 5xx rate > 1% for 5 minutes                        |
| High gRPC error rate | gRPC 5xx-equivalent status rate > 1% for 5 minutes |
| High latency         | P99 latency > 1s for 5 minutes                     |
| Pod restarts         | restart count increases within 15 minutes          |
| Kafka consumer lag   | consumer lag above agreed threshold                |
