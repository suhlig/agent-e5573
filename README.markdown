Agent E5573 collects some stats from the Huawei E5573 wireless router ("MiFi") and publishes it to an InfluxDB.

# Synopsis

```command
$ export INFLUXDB_PASSWORD=s3cret
$ agent-e5573 \
  --e5573-url http://192.168.8.1 \
  --influxdb-url https://influxdb.example.com \
  --influxdb-database e5573 \
  --influxdb-user example.com
```

`TODO` Run it with `--verbose` and it will print its the stats to `STDOUT`:

```command
$ $ agent-e5573 \
  --verbose \
  --e5573-url http://192.168.8.1 \
  --influxdb-url https://influxdb.example.com \
  --influxdb-database e5573 \
  --influxdb-user example.com
Timestamp: 2020-05-11T19:12:21+02:00
Battery: 1.00
WiFi Users: 1
Network Mode: 4G/LTE Enabled
Network Signal: 0.20
Connection Time: 49764 s
Downloaded: 93.64 MiB
Uploaded 98.70 MiB
```

`TODO` Run it without an InfluxDB URL, and it will just print the stats without attempting to write to an InfluxDB.

`TODO` Use `--json` to print the same information as a JSON struct:

```command
$ $ agent-e5573 --json --e5573-url http://192.168.8.1
{
  "timestamp": "2020-05-11T19:12:21+02:00",
  "system": {
    "battery": 1.00,
    "wifi-users": 1
  },
  "network": {
    "mode": "4G/LTE Enabled",
    "signal": 0.20
  },
  "traffic": {
    "connection-time": { "value": 49764, "unit": "s" },
    "downloaded": { "value": 93.64, "unit": "MiB" },
    "uploaded": { "value": 98.70, "unit": "MiB" }
  }
}
```

# System Design Choices

We want to publish the stats in regular intervals. The following choices come to mind:

1. systemd timers
1. cron
1. Keep it running and maintain a loop that publishes directly
1. New [Telegraf plugin](https://www.influxdata.com/blog/telegraf-go-collection-agent/)
1. Provide data to the [Telegraf exec plugin](https://community.influxdata.com/t/data-collection-question-best-way-to-feed-from-a-stats-catcher/11964)
1. Act as Prometheus exporter and use the [Telegraf Prometheus plugin](https://community.influxdata.com/t/own-telegraf-plugin-need-to-scrape-metrics-from-prometheus-clients/11878)

Since we do not publish more than once a minute, the granularity of systemd timers or cron seems to be sufficient.
