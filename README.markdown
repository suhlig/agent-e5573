`Agent E5573` collects some stats from the [Huawei E5573](https://en.wikipedia.org/wiki/Huawei_E5#Huawei_E5573) wireless router ("MiFi") and publishes them to an InfluxDB instance.

# Synopsis

```command
$ export INFLUXDB_PASSWORD=s3cret
$ agent-e5573 \
  --e5573-url http://192.168.8.1 \
  --influxdb-url https://influxdb.example.com \
  --influxdb-database e5573 \
  --influxdb-user example.com
```

Run it with `--verbose` and it will print its the stats to `STDOUT`:

```command
$ agent-e5573 \
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

Run it without an InfluxDB URL, and it will just print the stats without attempting to write to an InfluxDB.

# System Design Choices

We want to publish the stats in regular intervals. The following choices come to mind:

1. systemd timers
1. cron
1. Keep it running and maintain a loop that publishes directly
1. New [Telegraf plugin](https://www.influxdata.com/blog/telegraf-go-collection-agent/)
1. Provide data to the [Telegraf exec plugin](https://community.influxdata.com/t/data-collection-question-best-way-to-feed-from-a-stats-catcher/11964)
1. Act as Prometheus exporter and use the [Telegraf Prometheus plugin](https://community.influxdata.com/t/own-telegraf-plugin-need-to-scrape-metrics-from-prometheus-clients/11878)

Since we do not publish more than once a minute, the granularity of systemd timers or cron seems to be sufficient. And it helps keeping things simple.

# Development

Regular deployment is done with Ansible. For fast iteration, the following approach can be used:

1. Set the following environment variables:

    ```command
    $ export AGENT_E5573_HOST=pi.example.com # where the agent will be running
    $ export INFLUXDB_URL=https://influxdb.example.com:443 # where to push data
    $ export INFLUXDB_PASSWORD="t0ps3cr3t" # top secret
    ```

1. Optionally, if your target system is a Raspberry Pi <= 3, set the following variables on the build machine:

    ```command
    $ export GOOS=linux
    $ export GOARCH=arm
    $ export GOARM=7
    ```

1. Run the `setup` script once:

    ```command
    $ scripts/setup
    ```

1. Run this combination of scripts in order to `build`, `deploy` and `run` the agent:

    ```command
    $ scripts/build && scripts/deploy && scripts/run
    ```

    If you want to iterate quickly and build, deploy and run on each save, run `scripts/watch`.

1. Optionally, improve the code and supply a pull request.
